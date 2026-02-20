package inventory

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"ticketing-gozero/pkg/core/inventory/domain"
	"ticketing-gozero/pkg/core/inventory/partition"
	"ticketing-gozero/pkg/infra/inventory/event"
	"ticketing-gozero/pkg/infra/inventory/snapshot"
	"ticketing-gozero/pkg/infra/inventory/ttl"
	"ticketing-gozero/pkg/infra/inventory/wal"
)

type Service struct {
	logger         *slog.Logger
	partitionMgr   *partition.Manager
	walRepo        *wal.Repository
	snapshotRepo   *snapshot.Repository
	eventPublisher *event.Publisher
	holdStore      *ttl.Store

	walQueue             chan partition.MutationRecord
	snapshotInterval     time.Duration
	snapshotOpsThreshold int64
	opCounter            int64
}

type Config struct {
	ShardCount           int
	WALBuffer            int
	SnapshotInterval     time.Duration
	SnapshotOpsThreshold int64
}

type TryHoldInput struct {
	PartitionKey string
	HoldID       string
	Qty          int
	Capacity     int
}

type ReleaseInput struct {
	PartitionKey string
	HoldID       string
}

type ConfirmInput struct {
	PartitionKey string
	HoldID       string
}

func NewService(
	logger *slog.Logger,
	walRepo *wal.Repository,
	snapshotRepo *snapshot.Repository,
	eventPublisher *event.Publisher,
	holdStore *ttl.Store,
	cfg Config,
) *Service {
	if cfg.WALBuffer <= 0 {
		cfg.WALBuffer = 4096
	}
	if cfg.SnapshotInterval <= 0 {
		cfg.SnapshotInterval = 10 * time.Second
	}
	if cfg.SnapshotOpsThreshold <= 0 {
		cfg.SnapshotOpsThreshold = 500
	}
	walQueue := make(chan partition.MutationRecord, cfg.WALBuffer)
	return &Service{
		logger:               logger,
		partitionMgr:         partition.NewManager(cfg.ShardCount, walQueue),
		walRepo:              walRepo,
		snapshotRepo:         snapshotRepo,
		eventPublisher:       eventPublisher,
		holdStore:            holdStore,
		walQueue:             walQueue,
		snapshotInterval:     cfg.SnapshotInterval,
		snapshotOpsThreshold: cfg.SnapshotOpsThreshold,
	}
}

func (s *Service) Start(ctx context.Context) error {
	if err := s.Recover(ctx); err != nil {
		return err
	}
	go s.walWriterLoop(ctx)
	go s.snapshotLoop(ctx)
	go s.ttlReleaseLoop(ctx)
	return nil
}

func (s *Service) Recover(ctx context.Context) error {
	snapshots, err := s.snapshotRepo.LoadAll(ctx)
	if err != nil {
		return fmt.Errorf("load snapshots failed: %w", err)
	}
	snapshotSeq := map[string]int64{}
	for _, snap := range snapshots {
		if err := s.partitionMgr.RestoreState(ctx, snap.State); err != nil {
			return fmt.Errorf("restore snapshot state failed: %w", err)
		}
		snapshotSeq[snap.PartitionKey] = snap.SnapshotSeq
	}

	walRecords, err := s.walRepo.LoadAll(ctx)
	if err != nil {
		return fmt.Errorf("load wal failed: %w", err)
	}
	for _, rec := range walRecords {
		if rec.Seq <= snapshotSeq[rec.PartitionKey] {
			continue
		}
		if err := s.partitionMgr.ApplyRecoveredMutation(ctx, rec); err != nil {
			return fmt.Errorf("replay wal failed: %w", err)
		}
	}
	s.logger.Info("inventory recovery finished", "snapshot_count", len(snapshots), "wal_count", len(walRecords))
	return nil
}

func (s *Service) TryHold(ctx context.Context, in TryHoldInput) (*domain.PartitionState, error) {
	state, err := s.partitionMgr.TryHold(ctx, partition.TryHoldInput{
		PartitionKey: in.PartitionKey,
		HoldID:       in.HoldID,
		Qty:          in.Qty,
		Capacity:     in.Capacity,
	})
	if err != nil {
		return nil, err
	}
	if err := s.holdStore.Save(ctx, ttl.HoldValue{
		PartitionKey: in.PartitionKey,
		HoldID:       in.HoldID,
		Qty:          in.Qty,
	}); err != nil {
		_, _ = s.partitionMgr.ReleaseHold(ctx, partition.ReleaseInput{
			PartitionKey: in.PartitionKey,
			HoldID:       in.HoldID,
		})
		_ = s.holdStore.Remove(ctx, in.HoldID)
		return nil, fmt.Errorf("redis hold save failed: %w", err)
	}
	atomic.AddInt64(&s.opCounter, 1)
	return state, nil
}

func (s *Service) ReleaseHold(ctx context.Context, in ReleaseInput) (*domain.PartitionState, error) {
	state, err := s.partitionMgr.ReleaseHold(ctx, partition.ReleaseInput{
		PartitionKey: in.PartitionKey,
		HoldID:       in.HoldID,
	})
	if err != nil {
		return nil, err
	}
	_ = s.holdStore.Remove(ctx, in.HoldID)
	atomic.AddInt64(&s.opCounter, 1)
	return state, nil
}

func (s *Service) ConfirmHold(ctx context.Context, in ConfirmInput) (*domain.PartitionState, error) {
	state, err := s.partitionMgr.ConfirmHold(ctx, partition.ConfirmInput{
		PartitionKey: in.PartitionKey,
		HoldID:       in.HoldID,
	})
	if err != nil {
		return nil, err
	}
	_ = s.holdStore.Remove(ctx, in.HoldID)
	atomic.AddInt64(&s.opCounter, 1)
	return state, nil
}

func (s *Service) GetAvailability(ctx context.Context, partitionKey string) (int, bool, error) {
	return s.partitionMgr.GetAvailability(ctx, partitionKey)
}

func (s *Service) walWriterLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case rec := <-s.walQueue:
			saveCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			err := s.walRepo.Append(saveCtx, rec)
			cancel()
			if err != nil {
				s.logger.Error("wal append failed", "error", err, "partition_key", rec.PartitionKey, "seq", rec.Seq)
				continue
			}
			pubCtx, pubCancel := context.WithTimeout(ctx, 2*time.Second)
			err = s.eventPublisher.PublishMutation(pubCtx, rec)
			pubCancel()
			if err != nil {
				s.logger.Error("publish inventory event failed", "error", err, "partition_key", rec.PartitionKey, "seq", rec.Seq)
			}
		}
	}
}

func (s *Service) snapshotLoop(ctx context.Context) {
	ticker := time.NewTicker(s.snapshotInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if atomic.LoadInt64(&s.opCounter) >= s.snapshotOpsThreshold {
				if err := s.persistSnapshots(ctx); err != nil {
					s.logger.Error("periodic snapshot failed", "error", err)
				}
				atomic.StoreInt64(&s.opCounter, 0)
			}
		}
	}
}

func (s *Service) persistSnapshots(ctx context.Context) error {
	states, err := s.partitionMgr.ExportSnapshots(ctx)
	if err != nil {
		return err
	}
	for _, st := range states {
		rec := snapshot.Record{
			PartitionKey: st.PartitionKey,
			SnapshotSeq:  st.LastSeq,
			State:        st,
		}
		saveCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err := s.snapshotRepo.Upsert(saveCtx, rec)
		cancel()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ttlReleaseLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			expired, err := s.holdStore.PollExpired(ctx, 100)
			if err != nil {
				s.logger.Error("poll expired holds failed", "error", err)
				continue
			}
			for _, hold := range expired {
				_, err := s.ReleaseHold(ctx, ReleaseInput{
					PartitionKey: hold.PartitionKey,
					HoldID:       hold.HoldID,
				})
				if err != nil && !errors.Is(err, domain.ErrHoldNotFound) {
					s.logger.Error("release expired hold failed", "error", err, "hold_id", hold.HoldID)
					continue
				}
				_ = s.holdStore.Remove(ctx, hold.HoldID)
			}
		}
	}
}
