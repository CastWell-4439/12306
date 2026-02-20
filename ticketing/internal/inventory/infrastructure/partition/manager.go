package partition

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"ticketing/internal/inventory/domain"
)

type MutationRecord struct {
	PartitionKey string
	Seq          int64
	EventType    domain.EventType
	Payload      map[string]any
	OccurredAt   time.Time
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

type tryHoldCmd struct {
	in   TryHoldInput
	resp chan commandResult
}

type releaseCmd struct {
	in   ReleaseInput
	resp chan commandResult
}

type confirmCmd struct {
	in   ConfirmInput
	resp chan commandResult
}

type availabilityCmd struct {
	partitionKey string
	resp         chan availabilityResult
}

type restoreStateCmd struct {
	state *domain.PartitionState
	resp  chan error
}

type applyRecoveredMutationCmd struct {
	record MutationRecord
	resp   chan error
}

type exportSnapshotCmd struct {
	resp chan []*domain.PartitionState
}

type commandResult struct {
	state  *domain.PartitionState
	record *MutationRecord
	err    error
}

type availabilityResult struct {
	available int
	ok        bool
}

type shard struct {
	ch     chan any
	states map[string]*domain.PartitionState
}

type Manager struct {
	shards     []*shard
	walQueue   chan MutationRecord
	partitionN uint32
}

func NewManager(partitionN int, walQueue chan MutationRecord) *Manager {
	if partitionN <= 0 {
		partitionN = 32
	}
	m := &Manager{
		shards:     make([]*shard, 0, partitionN),
		walQueue:   walQueue,
		partitionN: uint32(partitionN),
	}
	for i := 0; i < partitionN; i++ {
		s := &shard{
			ch:     make(chan any, 1024),
			states: map[string]*domain.PartitionState{},
		}
		m.shards = append(m.shards, s)
		go s.loop(walQueue)
	}
	return m
}

func (m *Manager) TryHold(ctx context.Context, in TryHoldInput) (*domain.PartitionState, error) {
	if in.Qty <= 0 {
		return nil, domain.ErrInvalidQuantity
	}
	if in.PartitionKey == "" || in.HoldID == "" {
		return nil, fmt.Errorf("partition_key and hold_id are required")
	}
	resp := make(chan commandResult, 1)
	if err := m.send(ctx, in.PartitionKey, tryHoldCmd{in: in, resp: resp}); err != nil {
		return nil, err
	}
	res := <-resp
	return res.state, res.err
}

func (m *Manager) ReleaseHold(ctx context.Context, in ReleaseInput) (*domain.PartitionState, error) {
	resp := make(chan commandResult, 1)
	if err := m.send(ctx, in.PartitionKey, releaseCmd{in: in, resp: resp}); err != nil {
		return nil, err
	}
	res := <-resp
	return res.state, res.err
}

func (m *Manager) ConfirmHold(ctx context.Context, in ConfirmInput) (*domain.PartitionState, error) {
	resp := make(chan commandResult, 1)
	if err := m.send(ctx, in.PartitionKey, confirmCmd{in: in, resp: resp}); err != nil {
		return nil, err
	}
	res := <-resp
	return res.state, res.err
}

func (m *Manager) GetAvailability(ctx context.Context, partitionKey string) (int, bool, error) {
	resp := make(chan availabilityResult, 1)
	if err := m.send(ctx, partitionKey, availabilityCmd{partitionKey: partitionKey, resp: resp}); err != nil {
		return 0, false, err
	}
	out := <-resp
	return out.available, out.ok, nil
}

func (m *Manager) RestoreState(ctx context.Context, state *domain.PartitionState) error {
	resp := make(chan error, 1)
	if err := m.send(ctx, state.PartitionKey, restoreStateCmd{state: state, resp: resp}); err != nil {
		return err
	}
	return <-resp
}

func (m *Manager) ApplyRecoveredMutation(ctx context.Context, record MutationRecord) error {
	resp := make(chan error, 1)
	if err := m.send(ctx, record.PartitionKey, applyRecoveredMutationCmd{record: record, resp: resp}); err != nil {
		return err
	}
	return <-resp
}

func (m *Manager) ExportSnapshots(ctx context.Context) ([]*domain.PartitionState, error) {
	all := make([]*domain.PartitionState, 0)
	for idx := range m.shards {
		resp := make(chan []*domain.PartitionState, 1)
		if err := m.sendToShard(ctx, idx, exportSnapshotCmd{resp: resp}); err != nil {
			return nil, err
		}
		all = append(all, <-resp...)
	}
	return all, nil
}

func (m *Manager) send(ctx context.Context, partitionKey string, cmd any) error {
	return m.sendToShard(ctx, int(m.hash(partitionKey)%m.partitionN), cmd)
}

func (m *Manager) sendToShard(ctx context.Context, shardIndex int, cmd any) error {
	select {
	case m.shards[shardIndex].ch <- cmd:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *Manager) hash(key string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return h.Sum32()
}

func (s *shard) loop(walQueue chan MutationRecord) {
	for raw := range s.ch {
		switch cmd := raw.(type) {
		case tryHoldCmd:
			cmd.resp <- s.handleTryHold(cmd.in, walQueue)
		case releaseCmd:
			cmd.resp <- s.handleRelease(cmd.in, walQueue)
		case confirmCmd:
			cmd.resp <- s.handleConfirm(cmd.in, walQueue)
		case availabilityCmd:
			st, ok := s.states[cmd.partitionKey]
			if !ok {
				cmd.resp <- availabilityResult{available: 0, ok: false}
				break
			}
			cmd.resp <- availabilityResult{available: st.Available, ok: true}
		case restoreStateCmd:
			s.states[cmd.state.PartitionKey] = cloneState(cmd.state)
			cmd.resp <- nil
		case applyRecoveredMutationCmd:
			cmd.resp <- s.applyRecovered(cmd.record)
		case exportSnapshotCmd:
			states := make([]*domain.PartitionState, 0, len(s.states))
			for _, st := range s.states {
				states = append(states, cloneState(st))
			}
			cmd.resp <- states
		}
	}
}

func (s *shard) handleTryHold(in TryHoldInput, walQueue chan MutationRecord) commandResult {
	st := s.getOrInit(in.PartitionKey, in.Capacity)
	if st.Available < in.Qty {
		return commandResult{err: domain.ErrInsufficientStock}
	}
	if _, exists := st.Holds[in.HoldID]; exists {
		return commandResult{state: cloneState(st)}
	}
	if len(walQueue) >= cap(walQueue) {
		return commandResult{err: domain.ErrBackpressure}
	}

	st.Available -= in.Qty
	st.Holds[in.HoldID] = domain.Hold{HoldID: in.HoldID, Qty: in.Qty}
	st.LastSeq++

	rec := MutationRecord{
		PartitionKey: in.PartitionKey,
		Seq:          st.LastSeq,
		EventType:    domain.EventTypeHoldCreated,
		Payload: map[string]any{
			"hold_id":  in.HoldID,
			"qty":      in.Qty,
			"capacity": st.Capacity,
		},
		OccurredAt: time.Now().UTC(),
	}
	select {
	case walQueue <- rec:
		return commandResult{state: cloneState(st), record: &rec}
	default:
		// Roll back to preserve correctness when WAL cannot be accepted.
		delete(st.Holds, in.HoldID)
		st.Available += in.Qty
		st.LastSeq--
		return commandResult{err: domain.ErrBackpressure}
	}
}

func (s *shard) handleRelease(in ReleaseInput, walQueue chan MutationRecord) commandResult {
	st, ok := s.states[in.PartitionKey]
	if !ok {
		return commandResult{err: domain.ErrHoldNotFound}
	}
	hold, ok := st.Holds[in.HoldID]
	if !ok {
		return commandResult{err: domain.ErrHoldNotFound}
	}

	st.Available += hold.Qty
	delete(st.Holds, in.HoldID)
	st.LastSeq++

	rec := MutationRecord{
		PartitionKey: in.PartitionKey,
		Seq:          st.LastSeq,
		EventType:    domain.EventTypeHoldReleased,
		Payload: map[string]any{
			"hold_id": in.HoldID,
			"qty":     hold.Qty,
		},
		OccurredAt: time.Now().UTC(),
	}
	select {
	case walQueue <- rec:
		return commandResult{state: cloneState(st), record: &rec}
	default:
		// Roll back to preserve replayability when WAL cannot be accepted.
		st.LastSeq--
		st.Holds[in.HoldID] = hold
		st.Available -= hold.Qty
		return commandResult{err: domain.ErrBackpressure}
	}
}

func (s *shard) handleConfirm(in ConfirmInput, walQueue chan MutationRecord) commandResult {
	st, ok := s.states[in.PartitionKey]
	if !ok {
		return commandResult{err: domain.ErrHoldNotFound}
	}
	hold, ok := st.Holds[in.HoldID]
	if !ok {
		return commandResult{err: domain.ErrHoldNotFound}
	}

	delete(st.Holds, in.HoldID)
	st.Confirmed += hold.Qty
	st.LastSeq++

	rec := MutationRecord{
		PartitionKey: in.PartitionKey,
		Seq:          st.LastSeq,
		EventType:    domain.EventTypeHoldConfirmed,
		Payload: map[string]any{
			"hold_id": in.HoldID,
			"qty":     hold.Qty,
		},
		OccurredAt: time.Now().UTC(),
	}
	select {
	case walQueue <- rec:
		return commandResult{state: cloneState(st), record: &rec}
	default:
		// Roll back to preserve replayability when WAL cannot be accepted.
		st.LastSeq--
		st.Confirmed -= hold.Qty
		st.Holds[in.HoldID] = hold
		return commandResult{err: domain.ErrBackpressure}
	}
}

func (s *shard) applyRecovered(record MutationRecord) error {
	st := s.getOrInit(record.PartitionKey, intFromPayload(record.Payload, "capacity"))
	if record.Seq <= st.LastSeq {
		return nil
	}
	switch record.EventType {
	case domain.EventTypeHoldCreated:
		holdID := stringFromPayload(record.Payload, "hold_id")
		qty := intFromPayload(record.Payload, "qty")
		if _, exists := st.Holds[holdID]; !exists {
			st.Holds[holdID] = domain.Hold{HoldID: holdID, Qty: qty}
			st.Available -= qty
		}
	case domain.EventTypeHoldReleased:
		holdID := stringFromPayload(record.Payload, "hold_id")
		hold, ok := st.Holds[holdID]
		if ok {
			delete(st.Holds, holdID)
			st.Available += hold.Qty
		}
	case domain.EventTypeHoldConfirmed:
		holdID := stringFromPayload(record.Payload, "hold_id")
		hold, ok := st.Holds[holdID]
		if ok {
			delete(st.Holds, holdID)
			st.Confirmed += hold.Qty
		}
	}
	st.LastSeq = record.Seq
	return nil
}

func (s *shard) getOrInit(partitionKey string, capacity int) *domain.PartitionState {
	st, ok := s.states[partitionKey]
	if ok {
		return st
	}
	if capacity <= 0 {
		capacity = 100
	}
	st = domain.NewPartitionState(partitionKey, capacity)
	s.states[partitionKey] = st
	return st
}

func cloneState(in *domain.PartitionState) *domain.PartitionState {
	holds := make(map[string]domain.Hold, len(in.Holds))
	for k, v := range in.Holds {
		holds[k] = v
	}
	return &domain.PartitionState{
		PartitionKey: in.PartitionKey,
		Capacity:     in.Capacity,
		Available:    in.Available,
		Confirmed:    in.Confirmed,
		LastSeq:      in.LastSeq,
		Holds:        holds,
	}
}

func intFromPayload(payload map[string]any, key string) int {
	raw, ok := payload[key]
	if !ok {
		return 0
	}
	switch v := raw.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

func stringFromPayload(payload map[string]any, key string) string {
	raw, ok := payload[key]
	if !ok {
		return ""
	}
	v, _ := raw.(string)
	return v
}
