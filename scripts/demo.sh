#!/usr/bin/env sh
set -eu
BASE="${BASE_URL:-http://localhost:8080}"
curl -fsS -X POST "$BASE/v1/cells" -H 'Content-Type: application/json' -d '{"id":"cell-east","region":"ap-south-1","cluster":"cluster-east","capacity":8}' || true
curl -fsS -X POST "$BASE/v1/cells" -H 'Content-Type: application/json' -d '{"id":"cell-west","region":"ap-south-1","cluster":"cluster-west","capacity":5}' || true
curl -fsS -X POST "$BASE/v1/workloads" -H 'Content-Type: application/json' -d '{"id":"identity","name":"identity-api","replicas":3}' || true
curl -fsS -X POST "$BASE/v1/reconcile"
echo
curl -fsS "$BASE/v1/workloads"
echo
