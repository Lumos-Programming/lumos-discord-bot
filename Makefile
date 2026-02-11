INSTANCE ?= default
HOST     ?= 127.0.0.1
PORT     ?= 8080
GCLOUD   ?= gcloud

.PHONY: fs-ensure fs-start fs-stop fs-env

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
