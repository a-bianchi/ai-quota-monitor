# Compatibility with Go 1.22 on recent macOS.
ifeq ($(shell uname -s),Darwin)
GO_LDFLAGS ?= -linkmode=external
endif

.PHONY: test run build panel-password build-collector codex-login collect-codex collect-codex-once docker-up docker-down
test:
	cd server && go test -race -ldflags="$(GO_LDFLAGS)" ./...
run: build
	cd server && ./bin/ai-quota-monitor
build:
	cd server && go build -ldflags="$(GO_LDFLAGS)" -o bin/ai-quota-monitor ./cmd/server
ifeq ($(shell uname -s),Darwin)
	codesign --force --sign - server/bin/ai-quota-monitor
endif

# Asks for the panel password and prints the AUTH_PASSWORD_HASH line for server/.env.
panel-password: build
	@server/bin/ai-quota-monitor hash-password

docker-up:
	docker compose --env-file server/.env up --build -d
docker-down:
	docker compose --env-file server/.env down

ACCOUNT ?= codex-work
build-collector:
	cd server && go build -ldflags="$(GO_LDFLAGS)" -o bin/codex-collector ./cmd/collector
ifeq ($(shell uname -s),Darwin)
	codesign --force --sign - server/bin/codex-collector
endif
codex-login: build-collector
	cd server && ./bin/codex-collector -account "$(ACCOUNT)" -login
collect-codex: build-collector
	cd server && ./bin/codex-collector -account "$(ACCOUNT)"
collect-codex-once: build-collector
	cd server && ./bin/codex-collector -account "$(ACCOUNT)" -once
