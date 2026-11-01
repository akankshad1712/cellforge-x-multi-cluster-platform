# API

| Method | Path | Purpose |
|---|---|---|
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Readiness |
| GET | `/metrics` | Prometheus text metrics |
| GET | `/v1/cells` | List cells |
| POST | `/v1/cells` | Register cell |
| POST | `/v1/cells/{id}/state` | Update simulated health |
| GET | `/v1/workloads` | List workloads |
| POST | `/v1/workloads` | Submit workload intent |
| POST | `/v1/reconcile` | Run reconciliation |

Example cell: `{"id":"ap-south-a","region":"ap-south-1","cluster":"cluster-a","capacity":10}`

Example workload: `{"id":"catalog","name":"catalog-api","replicas":3}`
