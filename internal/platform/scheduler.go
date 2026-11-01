package platform

import "sort"

type Scheduler struct{ store *Store }

func NewScheduler(s *Store) *Scheduler { return &Scheduler{store: s} }

// ChooseCell uses deterministic best-fit placement among healthy cells with enough free capacity.
func (sc *Scheduler) ChooseCell(replicas int) (Cell, bool) {
	candidates := []Cell{}
	for _, c := range sc.store.Cells() {
		if c.State == CellHealthy && c.Capacity-c.Allocated >= replicas {
			candidates = append(candidates, c)
		}
	}
	if len(candidates) == 0 {
		return Cell{}, false
	}
	sort.Slice(candidates, func(i, j int) bool {
		ri := candidates[i].Capacity - candidates[i].Allocated - replicas
		rj := candidates[j].Capacity - candidates[j].Allocated - replicas
		if ri == rj {
			return candidates[i].ID < candidates[j].ID
		}
		return ri < rj
	})
	return candidates[0], true
}
func (sc *Scheduler) Reconcile() int {
	moved := 0
	for _, w := range sc.store.Workloads() {
		if w.State == WorkloadPending || w.State == WorkloadDisrupted {
			if c, ok := sc.ChooseCell(w.Replicas); ok && sc.store.Schedule(w.ID, c.ID) == nil {
				moved++
			}
		}
	}
	return moved
}
