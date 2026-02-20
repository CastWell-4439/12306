package query

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	segmentkafka "github.com/segmentio/kafka-go"
)

type fakeManualCommitConsumer struct {
	message     segmentkafka.Message
	fetchErr    error
	commitErr   error
	fetchCalls  int
	commitCalls int
}

func (f *fakeManualCommitConsumer) Fetch(_ context.Context) (segmentkafka.Message, error) {
	f.fetchCalls++
	if f.fetchErr != nil {
		return segmentkafka.Message{}, f.fetchErr
	}
	return f.message, nil
}

func (f *fakeManualCommitConsumer) Commit(_ context.Context, _ segmentkafka.Message) error {
	f.commitCalls++
	return f.commitErr
}

func newTestService() *Service {
	return &Service{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func TestConsumeOnce_HandlerErrorDoesNotCommit(t *testing.T) {
	t.Parallel()

	svc := newTestService()
	consumer := &fakeManualCommitConsumer{
		message: segmentkafka.Message{Value: []byte(`{"event_id":"evt-1"}`)},
	}

	keepRunning := svc.consumeOnce(
		context.Background(),
		consumer,
		"order.events",
		func(context.Context, []byte) error { return errors.New("boom") },
	)

	if !keepRunning {
		t.Fatal("consumeOnce unexpectedly requested stop")
	}
	if consumer.fetchCalls != 1 {
		t.Fatalf("expected fetch once, got %d", consumer.fetchCalls)
	}
	if consumer.commitCalls != 0 {
		t.Fatalf("expected no commit on handler failure, got %d", consumer.commitCalls)
	}
}

func TestConsumeOnce_SuccessCommitsOffset(t *testing.T) {
	t.Parallel()

	svc := newTestService()
	consumer := &fakeManualCommitConsumer{
		message: segmentkafka.Message{Value: []byte(`{"event_id":"evt-2"}`)},
	}

	keepRunning := svc.consumeOnce(
		context.Background(),
		consumer,
		"order.events",
		func(context.Context, []byte) error { return nil },
	)

	if !keepRunning {
		t.Fatal("consumeOnce unexpectedly requested stop")
	}
	if consumer.fetchCalls != 1 {
		t.Fatalf("expected fetch once, got %d", consumer.fetchCalls)
	}
	if consumer.commitCalls != 1 {
		t.Fatalf("expected one commit on success, got %d", consumer.commitCalls)
	}
}

