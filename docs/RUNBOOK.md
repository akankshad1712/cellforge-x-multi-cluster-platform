# Local incident runbook

1. Start the API with `make run`.
2. Register at least two cells with enough spare capacity.
3. Submit a workload and call `POST /v1/reconcile`.
4. Set its assigned cell to `offline` using `POST /v1/cells/{id}/state`.
5. Call `POST /v1/reconcile` again.
6. Inspect `GET /v1/workloads`, `GET /v1/cells`, and `GET /metrics`.

If no healthy cell has sufficient capacity, the workload remains disrupted. That is intentional: the scheduler must not invent capacity.
