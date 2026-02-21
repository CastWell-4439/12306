package ticket

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	commonkafka "ticketing-gozero/pkg/infra/kafka"
	"ticketing-gozero/pkg/core/ticket/domain"
	grpcclient "ticketing-gozero/pkg/infra/ticket/grpc_client"
	"ticketing-gozero/pkg/infra/ticket/outbox"
	"ticketing-gozero/pkg/infra/ticket/repository"
)

type eventProducer interface {
	Publish(ctx context.Context, topic string, key []byte, value []byte) error
}

type ticketOutboxStore interface {
	InsertTx(ctx context.Context, tx *sql.Tx, eventID string, aggregateID string, eventType string, payload map[string]any) error
	ListPending(ctx context.Context, limit int) ([]outbox.Event, error)
	MarkPublished(ctx context.Context, id int64) error
	MarkRetry(ctx context.Context, id int64, retryCount int, nextRetryAt time.Time, lastError string) error
}

type Worker struct {
	logger        *slog.Logger
	consumer      *commonkafka.Consumer
	producer      eventProducer
	repo          *repository.Repository
	outbox        ticketOutboxStore
	seatAllocator grpcclient.SeatAllocatorClient
}

type orderEventEnvelope struct {
	EventID     string         `json:"event_id"`
	AggregateID string         `json:"aggregate_id"`
	EventType   string         `json:"event_type"`
	OccurredAt  string         `json:"occurred_at"`
	Payload     map[string]any `json:"payload"`
}

func NewWorker(
	logger *slog.Logger,
	consumer *commonkafka.Consumer,
	producer eventProducer,
	repo *repository.Repository,
	outboxStore ticketOutboxStore,
	seatAllocator grpcclient.SeatAllocatorClient,
) *Worker {
	return &Worker{
		logger:        logger,
		consumer:      consumer,
		producer:      producer,
		repo:          repo,
		outbox:        outboxStore,
		seatAllocator: seatAllocator,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	go w.startOutboxPublisher(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		msg, err := w.consumer.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			w.logger.Error("read kafka message failed", "error", err)
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if err := w.handleMessage(ctx, msg.Value); err != nil {
			w.logger.Error("handle order event failed", "error", err)
		}
	}
}

func (w *Worker) handleMessage(ctx context.Context, raw []byte) error {
	var ev orderEventEnvelope
	if err := json.Unmarshal(raw, &ev); err != nil {
		return err
	}
	if ev.EventType != "OrderPaid" {
		return nil
	}

	orderID := ev.AggregateID
	seat, err := w.seatAllocator.AllocateSeat(ctx, orderID)
	if err != nil {
		return err
	}

	tx, err := w.repo.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ok, err := w.repo.IsOrderPaidTx(ctx, tx, orderID)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	inserted, err := w.repo.InsertTicketTx(ctx, tx, domain.Ticket{
		TicketID:      uuid.NewString(),
		OrderID:       orderID,
		PassengerName: seat,
	})
	if err != nil {
		return err
	}
	if inserted {
		if err := w.repo.MarkOrderTicketedTx(ctx, tx, orderID); err != nil {
			return err
		}
		eventID, payload := buildTicketIssuedEvent(orderID, seat)
		if err := w.outbox.InsertTx(ctx, tx, eventID, orderID, "TicketIssued", payload); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func buildTicketIssuedEvent(orderID string, seat string) (string, map[string]any) {
	eventID := uuid.NewString()
	return eventID, map[string]any{
		"event_id":     eventID,
		"aggregate_id": orderID,
		"event_type":   "TicketIssued",
		"occurred_at":  time.Now().UTC().Format(time.RFC3339Nano),
		"payload": map[string]any{
			"order_id": orderID,
			"seat_no":  seat,
		},
	}
}

func (w *Worker) startOutboxPublisher(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.publishOutboxBatch(ctx, 100)
		}
	}
}

func (w *Worker) publishOutboxBatch(ctx context.Context, limit int) {
	listCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	events, err := w.outbox.ListPending(listCtx, limit)
	cancel()
	if err != nil {
		w.logger.Error("load ticket outbox failed", "error", err)
		return
	}

	for _, ev := range events {
		raw, err := json.Marshal(ev.Payload)
		if err != nil {
			w.logger.Error("marshal ticket outbox payload failed", "error", err, "event_id", ev.EventID)
			w.markOutboxRetry(ctx, ev, err)
			continue
		}

		pubCtx, pubCancel := context.WithTimeout(ctx, 2*time.Second)
		err = w.producer.Publish(pubCtx, "ticket.events", []byte(ev.AggregateID), raw)
		pubCancel()
		if err != nil {
			w.logger.Error("publish ticket outbox failed", "error", err, "event_id", ev.EventID)
			w.markOutboxRetry(ctx, ev, err)
			continue
		}

		markCtx, markCancel := context.WithTimeout(ctx, 2*time.Second)
		err = w.outbox.MarkPublished(markCtx, ev.ID)
		markCancel()
		if err != nil {
			w.logger.Error("mark ticket outbox published failed", "error", err, "event_id", ev.EventID)
		}
	}
}

func (w *Worker) markOutboxRetry(ctx context.Context, ev outbox.Event, cause error) {
	retryCount := ev.RetryCount + 1
	nextRetryAt := time.Now().Add(retryBackoff(retryCount))
	errMsg := truncateError(cause, 240)

	markCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := w.outbox.MarkRetry(markCtx, ev.ID, retryCount, nextRetryAt, errMsg); err != nil {
		w.logger.Error("mark ticket outbox retry failed", "error", err, "event_id", ev.EventID)
	}
}

func retryBackoff(retryCount int) time.Duration {
	if retryCount <= 0 {
		return time.Second
	}
	if retryCount > 6 {
		retryCount = 6
	}
	return time.Duration(1<<uint(retryCount-1)) * time.Second
}

func truncateError(err error, maxLen int) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if len(msg) <= maxLen {
		return msg
	}
	return msg[:maxLen]
}


