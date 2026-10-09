# Local setup

**English** · [Español](local-setup.es.md)

The repo root contains `compose.yaml`. `server/configs/config.json` uses fictitious data; `server/configs/config.live.json` includes direct OpenRouter and direct personal/work Codex with login on the server.

Each account defines `id`, `name`, `provider`, `scope`, `source` and, for direct OpenRouter, `credentialEnv`. IDs are unique, made of lowercase letters, digits, hyphens and underscores. Strict JSON: unknown fields are rejected.

| Variable | Default value / use |
|---|---|
| `DEVICE_TOKEN` | Required, at least 32 characters; reading metrics |
| `COLLECTOR_TOKEN` | Optional; if set, at least 32 characters and different from the previous one |
| `AUTH_USERNAME`, `AUTH_PASSWORD_HASH` | Optional, set together; they enable the [web panel](panel.md). The hash comes from `make panel-password`; never a plain-text password |
| `MASTER_KEY` | Optional, 32+ characters; encrypts the provider credentials saved from the panel |
| `HTTP_HOST` | `127.0.0.1` without Docker; Compose uses `0.0.0.0` inside the container |
| `HTTP_PORT` | `8080`; in Compose it sets the host port |
| `HOST_BIND` | Compose only: `127.0.0.1` for local access |
| `CONFIG_NAME` | Go and Docker: file name inside `server/configs/`, default `config.json` |
| `MOCK_FILE` | Binary: `configs/mock-usage.json` from `server/` |
| `DATA_DIR` | Binary: `./data`; Compose: `/data` with a persistent volume |
| `POLL_INTERVAL` | `5m`; Go duration, minimum 1s |
| `STALE_AFTER` | `15m`; must be >= the interval |
| `OPENROUTER_PERSONAL_KEY` | Management key, only for the live configuration. Fallback: a key saved from the panel takes priority |
| `CURSOR_WORK_SESSION` | cursor.com session cookie, only for the live configuration; see [Cursor](providers/cursor.md) |

## Real OpenRouter with Compose

Edit `server/.env`: `CONFIG_NAME=config.live.json` and `OPENROUTER_PERSONAL_KEY` with your key. Then `make docker-up`. Codex accounts require completing the [server HTTP login](providers/codex.md).

To use your own configuration without versioning it: save `server/configs/config.local.json` and choose `CONFIG_NAME=config.local.json`. It works the same with Go and Docker, without changing mounts. `CONFIG_NAME` is the file name, not a path. `CONFIG_FILE` is no longer used.

The server loads `.env` from its working directory; `make run` runs from `server/`. You can also use `cd server && go run ./cmd/server` with a Go version compatible with your system. On macOS with Go 1.22 use `make run`, which takes care of the linker and signing.

`make docker-up` is equivalent to `docker compose --env-file server/.env up --build -d`, from the root. We use `--env-file` because Compose lives at the root and the secrets belong to the server.

Variables exported in the terminal or defined by Docker take priority over `.env`; if you exported CONFIG_NAME earlier, use `unset CONFIG_NAME` to pick up the value from the file. The dotenv loader supports simple assignments, quotes and full-line comments; it does not run commands or expand variables. Paste real values, not shell expressions.

The `.env` files and `server/data/` are ignored by Git. `server/.env.example` only carries empty placeholders. Do not paste tokens in issues. The server does not print keys or provider error bodies.

## Manual ingestion test

With `config.live.json` and a collector token defined, use a current UTC timestamp:

```sh
printf '{"updatedAt":"%s","quotas":[{"id":"5h","name":"5 horas","unit":"percent","used":11,"limit":100,"resetsAt":null}]}' \
  "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > /tmp/quota-report.json
curl -X PUT http://localhost:8080/accounts/codex-work/usage \
  -H 'Authorization: Bearer YOUR_COLLECTOR_TOKEN' \
  -H 'Content-Type: application/json' \
  --data-binary @/tmp/quota-report.json
```

This test sends fictitious numbers; it does not query Codex.

On macOS, `make build` uses the system linker and signs the binary locally (ad-hoc) for compatibility with Go 1.22. `make run` builds and starts that binary. No developer identity is used and no commercial signature is distributed.

Collector variables: `QUOTA_API_URL` (default http://localhost:8080), `CODEX_BIN` (default codex), `CODEX_POLL_INTERVAL` (default 5m, minimum 10s). Do not put ChatGPT credentials in `.env`.

`config.live.json` now uses source=direct for Codex: the ingestion test above requires an account with source=collector in a custom configuration. Do not send manual reports to direct accounts.
