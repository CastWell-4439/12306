package partition

import (
	"context"
	"errors"
	"testing"
	"time"

	"ticketing/internal/inventory/domain"
)

func TestReleaseHold_BackpressureRollsBackAndReturns(t *testing.T) {
	t.Parallel()

	walQueue := make(chan MutationRecord, 1)
	mgr := NewManager(1, walQueue)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := mgr.TryHold(ctx, TryHoldInput{
		PartitionKey: "p1",
		HoldID:       "h1",
		Qty:          2,
		Capacity:     10,
	})
	if err != nil {
		t.Fatalf("setup hold failed: %v", err)
	}

	// walQueue is already full now; release should fail fast without blocking.
	resultCh := make(chan error, 1)
	go func() {
		_, releaseErr := mgr.ReleaseHold(ctx, ReleaseInput{
			PartitionKey: "p1",
			HoldID:       "h1",
		})
		resultCh <- releaseErr
	}()

	select {
	case err = <-resultCh:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("ReleaseHold blocked under WAL backpressure")
	}

	if !errors.Is(err, domain.ErrBackpressure) {
		t.Fatalf("expected ErrBackpressure, got: %v", err)
	}

	states, err := mgr.ExportSnapshots(ctx)
	if err != nil {
		t.Fatalf("ExportSnapshots failed: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("expected 1 state, got %d", len(states))
	}
	st := states[0]
	if st.LastSeq != 1 {
		t.Fatalf("expected LastSeq=1 after rollback, got %d", st.LastSeq)
	}
	if st.Available != 8 {
		t.Fatalf("expected Available=8 after rollback, got %d", st.Available)
	}
	if len(st.Holds) != 1 {
		t.Fatalf("expected hold to remain after rollback, got %d holds", len(st.Holds))
	}
}

func TestConfirmHold_BackpressureRollsBackAndReturns(t *testing.T) {
	t.Parallel()

	walQueue := make(chan MutationRecord, 1)
	mgr := NewManager(1, walQueue)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := mgr.TryHold(ctx, TryHoldInput{
		PartitionKey: "p1",
		HoldID:       "h1",
		Qty:          3,
		Capacity:     10,
	})
	if err != nil {
		t.Fatalf("setup hold failed: %v", err)
	}

	// walQueue is already full now; confirm should fail fast without blocking.
	resultCh := make(chan error, 1)
	go func() {
		_, confirmErr := mgr.ConfirmHold(ctx, ConfirmInput{
			PartitionKey: "p1",
			HoldID:       "h1",
		})
		resultCh <- confirmErr
	}()

	select {
	case err = <-resultCh:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("ConfirmHold blocked under WAL backpressure")
	}

	if !errors.Is(err, domain.ErrBackpressure) {
		t.Fatalf("expected ErrBackpressure, got: %v", err)
	}

	states, err := mgr.ExportSnapshots(ctx)
	if err != nil {
		t.Fatalf("ExportSnapshots failed: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("expected 1 state, got %d", len(states))
	}
	st := states[0]
	if st.LastSeq != 1 {
		t.Fatalf("expected LastSeq=1 after rollback, got %d", st.LastSeq)
	}
	if st.Available != 7 {
		t.Fatalf("expected Available=7 after rollback, got %d", st.Available)
	}
	if st.Confirmed != 0 {
		t.Fatalf("expected Confirmed=0 after rollback, got %d", st.Confirmed)
	}
	if len(st.Holds) != 1 {
		t.Fatalf("expected hold to remain after rollback, got %d holds", len(st.Holds))
	}
}

