package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	commonkafka "ticketing/internal/common/kafka"
	"ticketing/internal/query/domain"
	"ticketing/internal/query/infrastructure/cache"
	"ticketing/internal/query/infrastructure/readmodel"

	segmentkafka "github.com/segmentio/kafka-go"
)

type Service struct {
	logger   *slog.Logger
	repo     *readmodel.Repository
	cache    *cache.Store
	consumer string
}

type eventEnvelope struct {
	EventID     string         `json:"event_id"`
	AggregateID string         `json:"aggregate_id"`
	EventType   string         `json:"event_type"`
	Payload     map[string]any `json:"payload"`
}

type manualCommitConsumer interface {
	Fetch(ctx context.Context) (segmentkafka.Message, error)
	Commit(ctx context.Context, msg segmentkafka.Message) error
}

func NewService(logger *slog.Logger, repo *readmodel.Repository, cacheStore *cache.Store) *Service {
	return &Service{
		logger:   logger,
		repo:     repo,
		cache:    cacheStore,
		consumer: "query-service",
	}
}

func (s *Service) GetOrderView(ctx context.Context, orderID string) (*domain.OrderView, error) {
	if cv, ok, err := s.cache.GetOrderView(ctx, orderID); err == nil && ok {
		return cv, nil
	}

	v, err := s.repo.GetOrderView(ctx, orderID)
	if err != nil {
		return nil, err
	}
	_ = s.cache.SetOrderView(ctx, v)
	return v, nil
}

func (s *Service) StartOrderEventsConsumer(ctx context.Context, c *commonkafka.Consumer) {
	s.consumeLoop(ctx, c, "order.events", s.handleOrderEvent)
}

func (s *Service) StartTicketEventsConsumer(ctx context.Context, c *commonkafka.Consumer) {
	s.consumeLoop(ctx, c, "ticket.events", s.handleTicketEvent)
}

func (s *Service) consumeLoop(
	ctx context.Context,
	c manualCommitConsumer,
	stream string,
	handler func(context.Context, []byte) error,
) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if !s.consumeOnce(ctx, c, stream, handler) {
			return
		}
	}
}

func (s *Service) consumeOnce(
	ctx context.Context,
	c manualCommitConsumer,
	stream string,
	handler func(context.Context, []byte) error,
) bool {
	msg, err := c.Fetch(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return false
		}
		s.logger.Error("query consumer fetch failed", "stream", stream, "error", err)
		time.Sleep(200 * time.Millisecond)
		return true
	}

	if err := handler(ctx, msg.Value); err != nil {
		// Do not commit on failure so the message can be retried.
		s.logger.Error("query consumer handle failed", "stream", stream, "error", err)
		return true
	}

	commitCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := c.Commit(commitCtx, msg); err != nil {
		if ctx.Err() != nil {
			return false
		}
		s.logger.Error("query consumer commit failed", "stream", stream, "error", err)
	}
	return true
}

func (s *Service) RebuildColdStart(ctx context.Context) error {
	return s.repo.RebuildFromOrders(ctx, 10000)
}

func (s *Service) handleOrderEvent(ctx context.Context, raw []byte) error {
	var ev eventEnvelope
	if err := json.Unmarshal(raw, &ev); err != nil {
		return err
	}
	tx, err := s.repo.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	inserted, err := s.repo.MarkConsumedTx(ctx, tx, ev.EventID, s.consumer)
	if err != nil || !inserted {
		return err
	}

	var (
		status        string
		amountCents   int64
		providerTxnID string
	)
	switch ev.EventType {
	case "OrderCreated":
		status = "INIT"
		amountCents = int64FromAny(ev.Payload["amount_cents"])
	case "OrderReserved":
		status = "RESERVED"
	case "OrderPaid":
		status = "PAID"
		providerTxnID = stringFromAny(ev.Payload["provider_txn_id"])
	case "OrderCancelled":
		status = "CANCELLED"
	default:
		return tx.Commit()
	}

	if err := s.repo.UpdateOrderFromOrderEventTx(ctx, tx, ev.AggregateID, status, amountCents, providerTxnID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	_ = s.cache.InvalidateOrderView(ctx, ev.AggregateID)
	return nil
}

func (s *Service) handleTicketEvent(ctx context.Context, raw []byte) error {
	var ev eventEnvelope
	if err := json.Unmarshal(raw, &ev); err != nil {
		return err
	}
	if ev.EventType != "TicketIssued" {
		return nil
	}
	tx, err := s.repo.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	inserted, err := s.repo.MarkConsumedTx(ctx, tx, ev.EventID, s.consumer)
	if err != nil || !inserted {
		return err
	}

	seatNo := stringFromAny(ev.Payload["seat_no"])
	if seatNo == "" {
		seatNo = stringFromAny(mapFromAny(ev.Payload["payload"])["seat_no"])
	}
	if err := s.repo.MarkTicketedTx(ctx, tx, ev.AggregateID, seatNo); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	_ = s.cache.InvalidateOrderView(ctx, ev.AggregateID)
	return nil
}

func int64FromAny(v any) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case float64:
		return int64(x)
	default:
		return 0
	}
}

func stringFromAny(v any) string {
	s, _ := v.(string)
	return s
}

func mapFromAny(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}
