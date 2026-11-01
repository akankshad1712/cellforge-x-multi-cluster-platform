package platform

import "time"

type CellState string

const (
	CellHealthy  CellState = "healthy"
	CellDegraded CellState = "degraded"
	CellOffline  CellState = "offline"
)

type WorkloadState string

const (
	WorkloadPending   WorkloadState = "pending"
	WorkloadScheduled WorkloadState = "scheduled"
	WorkloadDisrupted WorkloadState = "disrupted"
)

type Cell struct {
	ID        string    `json:"id"`
	Region    string    `json:"region"`
	Cluster   string    `json:"cluster"`
	Capacity  int       `json:"capacity"`
	Allocated int       `json:"allocated"`
	State     CellState `json:"state"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type Workload struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Replicas  int           `json:"replicas"`
	CellID    string        `json:"cellId,omitempty"`
	State     WorkloadState `json:"state"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}
type CreateCellRequest struct {
	ID       string `json:"id"`
	Region   string `json:"region"`
	Cluster  string `json:"cluster"`
	Capacity int    `json:"capacity"`
}
type CreateWorkloadRequest struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Replicas int    `json:"replicas"`
}
