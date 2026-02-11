INSTANCE ?= default
HOST     ?= 127.0.0.1
PORT     ?= 8080
GCLOUD   ?= gcloud

.PHONY: fs-ensure fs-start fs-stop fs-env test-fs

fs-ensure:
	@$(GCLOUD) components install cloud-firestore-emulator --quiet

fs-start: fs-ensure
	@echo "[fs-$(INSTANCE)] starting on $(HOST):$(PORT)"
	@$(GCLOUD) emulators firestore start --host-port=$(HOST):$(PORT)

fs-stop:
	@echo "[fs-$(INSTANCE)] stopping on port $(PORT)"
	@pids=$$(lsof -ti tcp:$(PORT) 2>/dev/null); \
	if [ -n "$$pids" ]; then kill $$pids; else echo "no process on $(PORT)"; fi

fs-env:
	@echo "FIRESTORE_EMULATOR_HOST=$(HOST):$(PORT)"

test-fs:
	@if ! lsof -ti tcp:$(PORT) >/dev/null 2>&1; then \
		echo "Firestore emulator is not running on $(HOST):$(PORT). Start it with: make fs-start"; \
		exit 1; \
	fi
	@FIRESTORE_EMULATOR_HOST="$(HOST):$(PORT)" FIRESTORE_PROJECT_ID="demo-test" go test -tags=integration ./...
