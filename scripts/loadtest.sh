#!/bin/bash
set -e

BASE_URL=${BASE_URL:-http://localhost:8080}
REQUESTS=${REQUESTS:-1000}
CONCURRENCY=${CONCURRENCY:-50}

if ! command -v hey &> /dev/null; then
    echo "Error: 'hey' is not installed."
    echo "Install with: go install github.com/rakyll/hey@latest"
    exit 1
fi

echo "Load testing $BASE_URL"
echo "Requests: $REQUESTS, Concurrency: $CONCURRENCY"
echo ""

echo "=== GET /health ==="
hey -n "$REQUESTS" -c "$CONCURRENCY" "$BASE_URL/health"

echo ""
echo "=== GET /v1/classi ==="
hey -n "$REQUESTS" -c "$CONCURRENCY" "$BASE_URL/v1/classi"

echo ""
echo "=== GET /v1/classi/guerriero ==="
hey -n "$REQUESTS" -c "$CONCURRENCY" "$BASE_URL/v1/classi/guerriero"

echo ""
echo "=== GET /v1/classi/guerriero/sotto-classi ==="
hey -n "$REQUESTS" -c "$CONCURRENCY" "$BASE_URL/v1/classi/guerriero/sotto-classi"
