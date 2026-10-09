# Troubleshooting

**English** · [Español](troubleshooting.es.md)

Check the logs first: the server explains why it does not start or which account failed.

- Local: the output of `make run`.
- Docker, from the repo root: `docker compose --env-file server/.env logs api`.

The logs do not include keys or provider response bodies.

## The server does not start

The log shows `startup failed` with the reason.

| Message | Cause and fix |
|---|---|
| `DEVICE_TOKEN must contain at least 32 characters` | `DEVICE_TOKEN` is missing from `server/.env` or too short. Generate it with `openssl rand -hex 32` |
| `COLLECTOR_TOKEN must contain at least 32 characters and differ from DEVICE_TOKEN` | Use a second token, different from the first |
| `set both AUTH_USERNAME and AUTH_PASSWORD_HASH, or neither` | Only one of the two panel variables is set |
| `AUTH_PASSWORD_HASH is not valid` | The hash is incomplete or altered. Generate it again with `make panel-password` and paste the whole line, without quotes |
| `AUTH_PASSWORD is no longer supported` | Remove `AUTH_PASSWORD` and use `AUTH_PASSWORD_HASH` |
| `MASTER_KEY must contain at least 32 characters` | Generate it with `openssl rand -base64 32` |
| `STALE_AFTER must be a duration >= POLL_INTERVAL` | Raise `STALE_AFTER` or lower `POLL_INTERVAL` |
| `invalid dotenv assignment at line N` | That line of `server/.env` is not of the form `NAME=value` |
| `direct openrouter requires credentialEnv`, `unknown provider`, `invalid or duplicate account ID/name` | Error in the accounts file. See [configuration](configuration.md#accounts) |
| `invalid or missing mock data for <id>` | A `mock` account has no entry in `server/configs/mock-usage.json` |
| `address already in use` | Another process uses the port. Change `HTTP_PORT` or stop the other process; `make run` and `make docker-up` cannot share 8080 |

With Compose, `Set DEVICE_TOKEN in .env` means the token is missing or `--env-file server/.env` was not passed. Use `make docker-up`.

## A change in `.env` has no effect

- Restart the server: variables are read only at startup.
- A variable exported in the terminal wins over the file. Try `unset VARIABLE_NAME`.
- `make run` runs from `server/`, so the file must be `server/.env`.

## API error responses

| Code | Meaning |
|---|---|
| 401 `unauthorized` | Wrong token or missing `Authorization: Bearer <token>` header. Check for stray spaces or quotes |
| 503 `endpoint disabled` | The route depends on something not configured: `COLLECTOR_TOKEN` for ingestion and token-based logins, or the panel user |
| 400 `invalid X-Timezone` | Use an IANA name, such as `America/Argentina/Buenos_Aires` or `UTC` |
| 404 on `PUT /accounts/{id}/usage` | The account does not exist or is not `source: collector` |
| 409 on ingestion | The report is older than the stored data |

## The panel

- **"El login del panel no está configurado en el servidor"** (panel login not configured): `AUTH_USERNAME` and `AUTH_PASSWORD_HASH` are missing, or you did not restart. See [create the user](panel.md#create-the-user).
- **Invalid username or password**: enter the password you typed in `make panel-password`, not the hash. If you forgot it, generate a new hash.
- **Too many failed attempts**: after 5 failures the login is blocked for 15 minutes.
- **"The server has no MASTER_KEY"**: a notice, not an error. Without `MASTER_KEY` credentials cannot be saved from the panel; environment variables keep working.
- **Saved credentials disappeared**: `MASTER_KEY` changed or the data folder was deleted. They have to be loaded again.

## An account shows no data

Query `GET /usage` and look at the account's `status`.

| `status` | What to do |
|---|---|
| `auth_required` | The credential is missing or expired. Repeat that provider's connection steps in the [guide](providers/connect.md) |
| `unavailable` | There was never a successful read. Check the logs, the server's Internet connection and, for Codex, that `CODEX_BIN` points to a valid executable |
| `stale` | There is data, but it is old. Wait one `POLL_INTERVAL` cycle; if it persists, check the logs |

Per provider:

- **Codex**: the login answers 502 if Codex CLI is not installed or the organization restricts device-code login. See [Codex](providers/codex.md).
- **Cursor**: the cookie expires and is not renewed automatically. Copy it again. See [Cursor](providers/cursor.md).
- **Claude**: if the code is rejected or 15 minutes passed, start the login again. See [Claude](providers/claude-subscription.md).
- **OpenRouter**: a management key is required; an inference key does not work.

## Docker and local network

- **Cannot connect from another machine on the network**: by default the port is published only on `127.0.0.1`. Set `HOST_BIND` to the server's LAN IP and run `make docker-up` again. See [deployment](deployment.md).
- **A new variable does not reach the container**: `compose.yaml` passes a fixed list of variables. Add it to its `environment` section.
- **Example data keeps showing up**: check `CONFIG_NAME`. Data is stored per `source` and `id`, so a `mock` account and a `direct` one do not share values.
- **Start from scratch**: `docker compose --env-file server/.env down -v` deletes the volume, including the Codex and Claude sessions.

## macOS

If `go build` or `go test` fail with linker errors, use `make build` and `make test`, which pass the required flags and sign the binary locally.
