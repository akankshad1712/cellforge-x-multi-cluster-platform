package platform

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBestFitChoosesTightestHealthyCell(t *testing.T) {
	s := NewStore()
	_ = s.AddCell(Cell{ID: "wide", Region: "r", Cluster: "c1", Capacity: 10, State: CellHealthy})
	_ = s.AddCell(Cell{ID: "tight", Region: "r", Cluster: "c2", Capacity: 5, State: CellHealthy})
	got, ok := NewScheduler(s).ChooseCell(3)
	if !ok || got.ID != "tight" {
		t.Fatalf("want tight cell, got %+v", got)
	}
}
func TestReconcilePlacesPendingWorkload(t *testing.T) {
	s := NewStore()
	_ = s.AddCell(Cell{ID: "a", Region: "r", Cluster: "c", Capacity: 8, State: CellHealthy})
	_ = s.AddWorkload(Workload{ID: "api", Name: "api", Replicas: 3, State: WorkloadPending})
	if n := NewScheduler(s).Reconcile(); n != 1 {
		t.Fatalf("want 1 placement, got %d", n)
	}
	if s.Workloads()[0].State != WorkloadScheduled {
		t.Fatal("workload not scheduled")
	}
}
func TestOfflineCellFailover(t *testing.T) {
	s := NewStore()
	for _, id := range []string{"a", "b"} {
		_ = s.AddCell(Cell{ID: id, Region: "r", Cluster: id, Capacity: 5, State: CellHealthy})
	}
	_ = s.AddWorkload(Workload{ID: "orders", Name: "orders", Replicas: 2, State: WorkloadPending})
	sc := NewScheduler(s)
	sc.Reconcile()
	old := s.Workloads()[0].CellID
	_ = s.SetCellState(old, CellOffline)
	_ = s.MarkDisrupted("orders")
	if n := sc.Reconcile(); n != 1 {
		t.Fatalf("want failover, got %d", n)
	}
	if s.Workloads()[0].CellID == old {
		t.Fatal("workload remained on offline cell")
	}
}
func TestRejectUnknownJSONField(t *testing.T) {
	s := NewStore()
	api := NewAPI(s, NewScheduler(s))
	req := httptest.NewRequest("POST", "/v1/cells", strings.NewReader(`{"id":"a","region":"r","cluster":"c","capacity":3,"extra":true}`))
	rec := httptest.NewRecorder()
	api.Handler().ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}
