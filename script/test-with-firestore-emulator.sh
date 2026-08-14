#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="$ROOT_DIR/docker/firestore-emulator/docker-compose.yml"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required"
  exit 1
fi

if docker compose version >/dev/null 2>&1; then
  DOCKER_COMPOSE=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
  DOCKER_COMPOSE=(docker-compose)
else
  echo "docker compose (v2) or docker-compose is required"
  exit 1
fi

"${DOCKER_COMPOSE[@]}" -f "$COMPOSE_FILE" up -d
trap '"${DOCKER_COMPOSE[@]}" -f "$COMPOSE_FILE" down -v' EXIT

echo "Waiting for Firestore emulator..."
for _ in $(seq 1 60); do
  if (echo >/dev/tcp/127.0.0.1/8080) >/dev/null 2>&1; then
    break
  fi
  sleep 0.5
done

export FIRESTORE_EMULATOR_HOST="127.0.0.1:8080"
export FIRESTORE_PROJECT_ID="demo-test"

cd "$ROOT_DIR"
go test -tags=integration ./...

