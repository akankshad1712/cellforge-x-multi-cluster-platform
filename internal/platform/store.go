package platform

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("resource already exists")
	ErrInvalid  = errors.New("invalid resource or placement constraint")
)

type Store struct {
	mu        sync.RWMutex
	cells     map[string]Cell
	workloads map[string]Workload
}

func NewStore() *Store { return &Store{cells: map[string]Cell{}, workloads: map[string]Workload{}} }
func (s *Store) AddCell(c Cell) error {
	if c.ID == "" || c.Region == "" || c.Cluster == "" || c.Capacity < 1 {
		return ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.cells[c.ID]; ok {
		return ErrConflict
	}
	c.UpdatedAt = time.Now().UTC()
	s.cells[c.ID] = c
	return nil
}
func (s *Store) AddWorkload(w Workload) error {
	if w.ID == "" || w.Name == "" || w.Replicas < 1 {
		return ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.workloads[w.ID]; ok {
		return ErrConflict
	}
	now := time.Now().UTC()
	w.CreatedAt = now
	w.UpdatedAt = now
	s.workloads[w.ID] = w
	return nil
}
func (s *Store) Cells() []Cell {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Cell, 0, len(s.cells))
	for _, v := range s.cells {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (s *Store) Workloads() []Workload {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Workload, 0, len(s.workloads))
	for _, v := range s.workloads {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (s *Store) SetCellState(id string, state CellState) error {
	if state != CellHealthy && state != CellDegraded && state != CellOffline {
		return ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.cells[id]
	if !ok {
		return ErrNotFound
	}
	c.State = state
	c.UpdatedAt = time.Now().UTC()
	s.cells[id] = c
	return nil
}
func (s *Store) Schedule(wid, cid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	w, ok := s.workloads[wid]
	if !ok {
		return ErrNotFound
	}
	c, ok := s.cells[cid]
	if !ok {
		return ErrNotFound
	}
	if c.State != CellHealthy || c.Allocated+w.Replicas > c.Capacity {
		return ErrInvalid
	}
	if w.CellID != "" && w.CellID != cid {
		old := s.cells[w.CellID]
		old.Allocated -= w.Replicas
		if old.Allocated < 0 {
			old.Allocated = 0
		}
		old.UpdatedAt = time.Now().UTC()
		s.cells[old.ID] = old
	}
	c.Allocated += w.Replicas
	c.UpdatedAt = time.Now().UTC()
	s.cells[cid] = c
	w.CellID = cid
	w.State = WorkloadScheduled
	w.UpdatedAt = time.Now().UTC()
	s.workloads[wid] = w
	return nil
}
func (s *Store) MarkDisrupted(wid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	w, ok := s.workloads[wid]
	if !ok {
		return ErrNotFound
	}
	if w.CellID != "" {
		c := s.cells[w.CellID]
		c.Allocated -= w.Replicas
		if c.Allocated < 0 {
			c.Allocated = 0
		}
		c.UpdatedAt = time.Now().UTC()
		s.cells[c.ID] = c
	}
	w.CellID = ""
	w.State = WorkloadDisrupted
	w.UpdatedAt = time.Now().UTC()
	s.workloads[wid] = w
	return nil
}
