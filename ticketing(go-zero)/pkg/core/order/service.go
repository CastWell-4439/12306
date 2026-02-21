package order

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"

	"ticketing-gozero/pkg/core/order/domain"
	"ticketing-gozero/pkg/infra/order/event"
	"ticketing-gozero/pkg/infra/order/inventory"
	"ticketing-gozero/pkg/infra/order/outbox"
	"ticketing-gozero/pkg/infra/order/repository"
)

var (
	ErrInvalidPaymentStatus = errors.New("invalid payment status")
	ErrInvalidSignature     = errors.New("invalid payment signature")
)

type Config struct {
	DefaultPartitionKey string
	DefaultHoldQty      int
	DefaultCapacity     int
	PaymentSignKey      string
}

type Service struct {
	logger          *slog.Logger
	repo            *repository.Repository
	outbox          *outbox.Repository
	publisher       *event.Publisher
	inventoryClient *inventory.Client
	cfg             Config
}

type CreateOrderInput struct {
	IdempotencyKey string
	AmountCents    int64
}

type ReserveOrderInput struct {
	OrderID      string
	PartitionKey string
	HoldID       string
	Qty          int
	Capacity     int
}

type PaymentCallbackInput struct {
	OrderID       string
	ProviderTxnID string
	Status        string
	PartitionKey  string
	HoldID        string
	Signature     string
}

type CancelOrderInput struct {
	OrderID      string
	PartitionKey string
	HoldID       string
}

func NewService(
	logger *slog.Logger,
	repo *repository.Repository,
	outboxRepo *outbox.Repository,
	publisher *event.Publisher,
	inventoryClient *inventory.Client,
	cfg Config,
) *Service {
	if cfg.DefaultPartitionKey == "" {
		cfg.DefaultPartitionKey = "G123|2026-02-11|2nd"
	}
	if cfg.DefaultHoldQty <= 0 {
		cfg.DefaultHoldQty = 1
	}
	if cfg.DefaultCapacity <= 0 {
		cfg.DefaultCapacity = 500
	}
	return &Service{
		logger:          logger,
		repo:            repo,
		outbox:          outboxRepo,
		publisher:       publisher,
		inventoryClient: inventoryClient,
		cfg:             cfg,
	}
}

func (s *Service) CreateOrder(ctx context.Context, in CreateOrderInput) (*domain.Order, error) {
	existing, err := s.repo.FindByIdempotencyKey(ctx, in.IdempotencyKey)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, domain.ErrOrderNotFound) {
		return nil, err
	}

	order, err := domain.NewOrder(uuid.NewString(), in.IdempotencyKey, in.AmountCents)
	if err != nil {
		return nil, err
	}

	tx, err := s.repo.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.repo.InsertTx(ctx, tx, order); err != nil {
		if stringsHasDuplicate(err) {
			existingOrder, e := s.repo.FindByIdempotencyKey(ctx, in.IdempotencyKey)
			if e == nil {
				return existingOrder, nil
			}
		}
		return nil, err
	}

	if err := s.outbox.InsertTx(ctx, tx, uuid.NewString(), order.OrderID, "OrderCreated", map[string]any{
		"order_id":        order.OrderID,
		"idempotency_key": order.IdempotencyKey,
		"status":          order.Status,
		"amount_cents":    order.AmountCents,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, order.OrderID)
}

func (s *Service) ReserveOrder(ctx context.Context, in ReserveOrderInput) (*domain.Order, error) {
	current, err := s.repo.FindByID(ctx, in.OrderID)
	if err != nil {
		return nil, err
	}
	if current.Status == domain.StatusReserved {
		return current, nil
	}
	if current.Status != domain.StatusInit {
		return nil, domain.ErrInvalidStateTransfer
	}

	partitionKey, holdID, qty, capacity := s.resolveHoldConfig(in.OrderID, in.PartitionKey, in.HoldID, in.Qty, in.Capacity)
	if err := s.inventoryClient.TryHold(ctx, inventory.TryHoldInput{
		PartitionKey: partitionKey,
		HoldID:       holdID,
		Qty:          qty,
		Capacity:     capacity,
	}); err != nil {
		return nil, err
	}
	shouldCompensateRelease := true
	defer func() {
		if !shouldCompensateRelease {
			return
		}
		releaseCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if releaseErr := s.inventoryClient.ReleaseHold(releaseCtx, inventory.ReleaseInput{
			PartitionKey: partitionKey,
			HoldID:       holdID,
		}); releaseErr != nil && !errors.Is(releaseErr, inventory.ErrHoldNotFound) {
			s.logger.Error("reserve compensation release failed", "order_id", in.OrderID, "error", releaseErr)
		}
	}()

	tx, err := s.repo.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ok, err := s.repo.UpdateStatusTx(ctx, tx, in.OrderID, domain.StatusInit, domain.StatusReserved)
	if err != nil {
		return nil, err
	}
	if !ok {
		current, qErr := s.repo.FindByID(ctx, in.OrderID)
		if qErr != nil {
			return nil, qErr
		}
		if current.Status == domain.StatusReserved {
			shouldCompensateRelease = false
			return current, nil
		}
		return nil, domain.ErrInvalidStateTransfer
	}

	if err := s.outbox.InsertTx(ctx, tx, uuid.NewString(), in.OrderID, "OrderReserved", map[string]any{
		"order_id":      in.OrderID,
		"status":        domain.StatusReserved,
		"partition_key": partitionKey,
		"hold_id":       holdID,
		"hold_qty":      qty,
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	shouldCompensateRelease = false
	return s.repo.FindByID(ctx, in.OrderID)
}

func (s *Service) PaymentCallback(ctx context.Context, in PaymentCallbackInput) (*domain.Order, error) {
	if err := s.verifyPaymentSignature(in); err != nil {
		return nil, err
	}
	if !strings.EqualFold(in.Status, "SUCCESS") {
		return nil, ErrInvalidPaymentStatus
	}

	partitionKey, holdID, _, _ := s.resolveHoldConfig(in.OrderID, in.PartitionKey, in.HoldID, 0, 0)
	current, err := s.repo.FindByID(ctx, in.OrderID)
	if err != nil {
		return nil, err
	}
	if current.Status == domain.StatusPaid || current.Status == domain.StatusTicketed {
		return current, nil
	}
	if current.Status != domain.StatusReserved {
		return nil, domain.ErrInvalidStateTransfer
	}

	if err := s.inventoryClient.ConfirmHold(ctx, inventory.ConfirmInput{
		PartitionKey: partitionKey,
		HoldID:       holdID,
	}); err != nil && !errors.Is(err, inventory.ErrHoldNotFound) {
		return nil, err
	}

	tx, err := s.repo.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.repo.InsertPaymentTx(ctx, tx, uuid.NewString(), in.OrderID, in.ProviderTxnID, in.Status); err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			existing, findErr := s.repo.FindByID(ctx, in.OrderID)
			if findErr != nil {
				return nil, findErr
			}
			return existing, nil
		}
		return nil, err
	}

	ok, err := s.repo.UpdateStatusTx(ctx, tx, in.OrderID, domain.StatusReserved, domain.StatusPaid)
	if err != nil {
		return nil, err
	}
	if !ok {
		current, qErr := s.repo.FindByID(ctx, in.OrderID)
		if qErr != nil {
			return nil, qErr
		}
		if current.Status == domain.StatusPaid {
			return current, nil
		}
		return nil, domain.ErrInvalidStateTransfer
	}

	if err := s.outbox.InsertTx(ctx, tx, uuid.NewString(), in.OrderID, "OrderPaid", map[string]any{
		"order_id":        in.OrderID,
		"provider_txn_id": in.ProviderTxnID,
		"partition_key":   partitionKey,
		"hold_id":         holdID,
		"status":          domain.StatusPaid,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, in.OrderID)
}

func (s *Service) CancelOrder(ctx context.Context, in CancelOrderInput) (*domain.Order, error) {
	current, err := s.repo.FindByID(ctx, in.OrderID)
	if err != nil {
		return nil, err
	}
	if current.Status == domain.StatusCancelled {
		return current, nil
	}
	if current.Status != domain.StatusInit && current.Status != domain.StatusReserved {
		return nil, domain.ErrInvalidStateTransfer
	}

	partitionKey, holdID, _, _ := s.resolveHoldConfig(in.OrderID, in.PartitionKey, in.HoldID, 0, 0)
	if current.Status == domain.StatusReserved {
		if err := s.inventoryClient.ReleaseHold(ctx, inventory.ReleaseInput{
			PartitionKey: partitionKey,
			HoldID:       holdID,
		}); err != nil && !errors.Is(err, inventory.ErrHoldNotFound) {
			return nil, err
		}
	}

	tx, err := s.repo.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ok, err := s.repo.UpdateStatusTx(ctx, tx, in.OrderID, current.Status, domain.StatusCancelled)
	if err != nil {
		return nil, err
	}
	if !ok {
		latest, qErr := s.repo.FindByID(ctx, in.OrderID)
		if qErr != nil {
			return nil, qErr
		}
		if latest.Status == domain.StatusCancelled {
			return latest, nil
		}
		return nil, domain.ErrInvalidStateTransfer
	}

	if err := s.outbox.InsertTx(ctx, tx, uuid.NewString(), in.OrderID, "OrderCancelled", map[string]any{
		"order_id":      in.OrderID,
		"partition_key": partitionKey,
		"hold_id":       holdID,
		"status":        domain.StatusCancelled,
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, in.OrderID)
}

func (s *Service) StartOutboxPublisher(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			batchCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			events, err := s.outbox.ListPending(batchCtx, 100)
			cancel()
			if err != nil {
				s.logger.Error("load pending outbox failed", "error", err)
				continue
			}
			for _, ev := range events {
				pubCtx, pubCancel := context.WithTimeout(ctx, 2*time.Second)
				err := s.publisher.Publish(pubCtx, ev)
				pubCancel()
				if err != nil {
					s.logger.Error("publish outbox failed", "error", err, "event_id", ev.EventID)
					continue
				}
				markCtx, markCancel := context.WithTimeout(ctx, 2*time.Second)
				err = s.outbox.MarkPublished(markCtx, ev.ID)
				markCancel()
				if err != nil {
					s.logger.Error("mark outbox published failed", "error", err, "event_id", ev.EventID)
				}
			}
		}
	}
}

func stringsHasDuplicate(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == 1062
}

func (s *Service) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("query order failed: %w", err)
	}
	return order, nil
}

func (s *Service) resolveHoldConfig(orderID string, partitionKey string, holdID string, qty int, capacity int) (string, string, int, int) {
	resolvedPartition := strings.TrimSpace(partitionKey)
	if resolvedPartition == "" {
		resolvedPartition = s.cfg.DefaultPartitionKey
	}
	resolvedHoldID := strings.TrimSpace(holdID)
	if resolvedHoldID == "" {
		resolvedHoldID = orderID
	}
	resolvedQty := qty
	if resolvedQty <= 0 {
		resolvedQty = s.cfg.DefaultHoldQty
	}
	if resolvedQty <= 0 {
		resolvedQty = 1
	}
	resolvedCapacity := capacity
	if resolvedCapacity <= 0 {
		resolvedCapacity = s.cfg.DefaultCapacity
	}
	if resolvedCapacity <= 0 {
		resolvedCapacity = 500
	}
	return resolvedPartition, resolvedHoldID, resolvedQty, resolvedCapacity
}

func (s *Service) verifyPaymentSignature(in PaymentCallbackInput) error {
	secret := strings.TrimSpace(s.cfg.PaymentSignKey)
	if secret == "" {
		return nil
	}
	provided := strings.ToLower(strings.TrimSpace(in.Signature))
	if provided == "" {
		return ErrInvalidSignature
	}
	payload := fmt.Sprintf("%s|%s|%s", in.OrderID, in.ProviderTxnID, strings.ToUpper(strings.TrimSpace(in.Status)))
	expected := signHMACSHA256(secret, payload)
	if !hmac.Equal([]byte(provided), []byte(expected)) {
		return ErrInvalidSignature
	}
	return nil
}

func signHMACSHA256(secret string, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}


