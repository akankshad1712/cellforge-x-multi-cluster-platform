package platform

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
)

type API struct {
	store     *Store
	scheduler *Scheduler
	requests  atomic.Uint64
}

func NewAPI(s *Store, sc *Scheduler) *API { return &API{store: s, scheduler: sc} }
func (a *API) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]string{"status": "ok"}) })
	m.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]string{"status": "ready"}) })
	m.HandleFunc("GET /metrics", a.metrics)
	m.HandleFunc("GET /v1/cells", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, a.store.Cells()) })
	m.HandleFunc("POST /v1/cells", a.createCell)
	m.HandleFunc("POST /v1/cells/{id}/state", a.setCellState)
	m.HandleFunc("GET /v1/workloads", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, a.store.Workloads()) })
	m.HandleFunc("POST /v1/workloads", a.createWorkload)
	m.HandleFunc("POST /v1/reconcile", a.reconcile)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.requests.Add(1)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		m.ServeHTTP(w, r)
	})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON request"})
		return false
	}
	return true
}
func (a *API) createCell(w http.ResponseWriter, r *http.Request) {
	var q CreateCellRequest
	if !decode(w, r, &q) {
		return
	}
	if e := a.store.AddCell(Cell{ID: q.ID, Region: q.Region, Cluster: q.Cluster, Capacity: q.Capacity, State: CellHealthy}); e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, 201, map[string]string{"created": q.ID})
}
func (a *API) createWorkload(w http.ResponseWriter, r *http.Request) {
	var q CreateWorkloadRequest
	if !decode(w, r, &q) {
		return
	}
	if e := a.store.AddWorkload(Workload{ID: q.ID, Name: q.Name, Replicas: q.Replicas, State: WorkloadPending}); e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, 201, map[string]string{"created": q.ID, "state": "pending"})
}
func (a *API) setCellState(w http.ResponseWriter, r *http.Request) {
	var q struct {
		State CellState `json:"state"`
	}
	if !decode(w, r, &q) {
		return
	}
	if e := a.store.SetCellState(r.PathValue("id"), q.State); e != nil {
		writeError(w, e)
		return
	}
	writeJSON(w, 200, map[string]string{"cell": r.PathValue("id"), "state": string(q.State)})
}
func (a *API) reconcile(w http.ResponseWriter, r *http.Request) {
	for _, c := range a.store.Cells() {
		if c.State != CellHealthy {
			for _, wl := range a.store.Workloads() {
				if wl.CellID == c.ID {
					_ = a.store.MarkDisrupted(wl.ID)
				}
			}
		}
	}
	n := a.scheduler.Reconcile()
	writeJSON(w, 200, map[string]any{"reconciled": n, "workloads": a.store.Workloads()})
}
func (a *API) metrics(w http.ResponseWriter, r *http.Request) {
	cells := a.store.Cells()
	workloads := a.store.Workloads()
	healthy, capacity, allocated, scheduled := 0, 0, 0, 0
	for _, c := range cells {
		if c.State == CellHealthy {
			healthy++
		}
		capacity += c.Capacity
		allocated += c.Allocated
	}
	for _, v := range workloads {
		if v.State == WorkloadScheduled {
			scheduled++
		}
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = fmt.Fprintf(w, "cellforge_cells %d\ncellforge_healthy_cells %d\ncellforge_capacity_units %d\ncellforge_allocated_units %d\ncellforge_scheduled_workloads %d\ncellforge_http_requests_total %d\n", len(cells), healthy, capacity, allocated, scheduled, a.requests.Load())
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, e error) {
	status := 400
	if errors.Is(e, ErrConflict) {
		status = 409
	}
	if errors.Is(e, ErrNotFound) {
		status = 404
	}
	if errors.Is(e, ErrInvalid) {
		status = 422
	}
	writeJSON(w, status, map[string]string{"error": e.Error()})
}
