# Local setup

**English** · [Español](local-setup.es.md)

How to run the server on your machine without Docker. For Docker Compose, see [deployment](deployment.md).

## Requirements

- Git.
- Go 1.22 or later.
- `make` and `openssl`.
- On macOS, Xcode Command Line Tools: `make build` uses the system linker and signs the binary locally (ad-hoc signature, no developer identity).
- Codex CLI, only if you are going to use Codex accounts. Set it with `CODEX_BIN`.

The server uses only the Go standard library: there are no dependencies to download.

## First run

Every command runs from the repo root.

1. Create the environment file:

   ```sh
   cp server/.env.example server/.env
   ```

2. Generate a token and paste it into `DEVICE_TOKEN` in `server/.env`:

   ```sh
   openssl rand -hex 32
   ```

3. Build and start:

   ```sh
   make run
   ```

4. In another terminal, check that it responds:

   ```sh
   curl http://localhost:8080/ping
   curl -H 'Authorization: Bearer YOUR_DEVICE_TOKEN' http://localhost:8080/usage
   ```

The server listens on `127.0.0.1:8080` and starts in example mode: six fictitious accounts with `source: mock`. `used` is the consumption; the balance is `limit - used`.

Data is stored in `server/data/`, which is in `.gitignore` together with `server/.env`.

## `make` commands

| Command | What it does |
|---|---|
| `make run` | Builds and starts the server from `server/` |
| `make build` | Builds `server/bin/ai-quota-monitor` |
| `make test` | Runs the tests with the race detector |
| `make panel-password` | Asks for a password and prints the `AUTH_PASSWORD_HASH=...` line |
| `make docker-up` / `make docker-down` | Starts or stops the service with Docker Compose |
| `make build-collector` | Builds the optional Codex collector |

You can also use `cd server && go run ./cmd/server`. On macOS prefer `make run`, which takes care of the linker and signing.

## Next steps

- **Web panel**: create the user with `make panel-password` and open <http://localhost:8080/>. See [web panel](panel.md).
- **Real accounts**: set `CONFIG_NAME=config.live.json` in `server/.env` and follow the [provider guide](providers/connect.md).
- **Variables and accounts file**: see [configuration](configuration.md).

## Testing ingestion by hand

Ingestion only accepts accounts with `source: collector`, and the bundled configurations have none. To test it:

1. Save `server/configs/config.local.json` with an account of that kind:

   ```json
   {"accounts": [{"id": "codex-work", "name": "Codex work", "provider": "openai", "scope": "work", "source": "collector"}]}
   ```

2. In `server/.env`, set `CONFIG_NAME=config.local.json` and a `COLLECTOR_TOKEN` generated with `openssl rand -hex 32`. Restart with `make run`.

3. Send a report with the current UTC time:

   ```sh
   printf '{"updatedAt":"%s","quotas":[{"id":"5h","name":"5 horas","unit":"percent","used":11,"limit":100,"resetsAt":null}]}' \
     "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > /tmp/quota-report.json
   curl -X PUT http://localhost:8080/accounts/codex-work/usage \
     -H 'Authorization: Bearer YOUR_COLLECTOR_TOKEN' \
     -H 'Content-Type: application/json' \
     --data-binary @/tmp/quota-report.json
   ```

These are fictitious numbers: this test does not query Codex. Do not send reports to `direct` accounts; the server rejects them with 404.

## Optional Codex collector

`server/cmd/collector` reads the Codex quotas on your machine and sends them to the API. It is only needed if you declare an `openai` account with `source: collector`; with `config.live.json` the server queries Codex directly and it is not used.

```sh
make codex-login ACCOUNT=codex-work     # signs in to an isolated profile
make collect-codex-once ACCOUNT=codex-work
make collect-codex ACCOUNT=codex-work   # sends every CODEX_POLL_INTERVAL
```

Its variables are in [configuration](configuration.md#optional-collector). Do not put ChatGPT credentials in `.env`.
