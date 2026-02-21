package ticket

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"ticketing-gozero/pkg/infra/ticket/outbox"
)

type fakeProducer struct {
	publishErr   error
	publishCalls int
}

func (f *fakeProducer) Publish(_ context.Context, _ string, _ []byte, _ []byte) error {
	f.publishCalls++
	return f.publishErr
}

type retryMark struct {
	id         int64
	retryCount int
	nextRetry  time.Time
	lastErr    string
}

type fakeOutboxStore struct {
	events          []outbox.Event
	markedPublished []int64
	markedRetry     []retryMark
}

func (f *fakeOutboxStore) InsertTx(_ context.Context, _ *sql.Tx, _ string, _ string, _ string, _ map[string]any) error {
	return nil
}

func (f *fakeOutboxStore) ListPending(_ context.Context, _ int) ([]outbox.Event, error) {
	return f.events, nil
}

func (f *fakeOutboxStore) MarkPublished(_ context.Context, id int64) error {
	f.markedPublished = append(f.markedPublished, id)
	return nil
}

func (f *fakeOutboxStore) MarkRetry(_ context.Context, id int64, retryCount int, nextRetryAt time.Time, lastError string) error {
	f.markedRetry = append(f.markedRetry, retryMark{
		id:         id,
		retryCount: retryCount,
		nextRetry:  nextRetryAt,
		lastErr:    lastError,
	})
	return nil
}

func testWorkerWithOutbox(prod *fakeProducer, store *fakeOutboxStore) *Worker {
	return &Worker{
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		producer: prod,
		outbox:   store,
	}
}

func TestPublishOutboxBatch_MarkPublishedOnSuccess(t *testing.T) {
	t.Parallel()

	prod := &fakeProducer{}
	store := &fakeOutboxStore{
		events: []outbox.Event{
			{
				ID:          1,
				EventID:     "evt-1",
				AggregateID: "order-1",
				Payload: map[string]any{
					"event_id": "evt-1",
				},
			},
		},
	}
	worker := testWorkerWithOutbox(prod, store)

	worker.publishOutboxBatch(context.Background(), 10)

	if prod.publishCalls != 1 {
		t.Fatalf("expected one publish call, got %d", prod.publishCalls)
	}
	if len(store.markedPublished) != 1 || store.markedPublished[0] != 1 {
		t.Fatalf("expected event 1 to be marked published, got %+v", store.markedPublished)
	}
	if len(store.markedRetry) != 0 {
		t.Fatalf("expected no retry marks, got %+v", store.markedRetry)
	}
}

func TestPublishOutboxBatch_MarkRetryOnPublishFailure(t *testing.T) {
	t.Parallel()

	prod := &fakeProducer{publishErr: errors.New("kafka down")}
	store := &fakeOutboxStore{
		events: []outbox.Event{
			{
				ID:          2,
				EventID:     "evt-2",
				AggregateID: "order-2",
				RetryCount:  2,
				Payload: map[string]any{
					"event_id": "evt-2",
				},
			},
		},
	}
	worker := testWorkerWithOutbox(prod, store)

	before := time.Now()
	worker.publishOutboxBatch(context.Background(), 10)

	if len(store.markedPublished) != 0 {
		t.Fatalf("expected no published marks on failure, got %+v", store.markedPublished)
	}
	if len(store.markedRetry) != 1 {
		t.Fatalf("expected one retry mark, got %+v", store.markedRetry)
	}
	got := store.markedRetry[0]
	if got.id != 2 {
		t.Fatalf("expected retry id=2, got %d", got.id)
	}
	if got.retryCount != 3 {
		t.Fatalf("expected retry count to increment to 3, got %d", got.retryCount)
	}
	if !got.nextRetry.After(before) {
		t.Fatalf("expected next retry time after now, got %s", got.nextRetry)
	}
	if got.lastErr == "" {
		t.Fatal("expected last error to be recorded")
	}
}


