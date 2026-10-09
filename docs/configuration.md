# Configuration

**English** · [Español](configuration.es.md)

The server is configured with two things:

- **Environment variables**, in `server/.env`: tokens, secrets and process settings.
- **An accounts file**, in `server/configs/`: which services are monitored and whether they are personal or work.

Restart the server after changing either of them.

## Environment variables

Copy the example file from the repo root and fill it in:

```sh
cp server/.env.example server/.env
```

The server reads `.env` from its working directory; `make run` runs it from `server/`. `make docker-up` passes the same file to Compose. The file is in `.gitignore`.

### Tokens and access

| Variable | Required | Purpose | Example |
|---|---|---|---|
| `DEVICE_TOKEN` | Yes | Read token for `GET /usage`; the one the display uses. At least 32 characters | Output of `openssl rand -hex 32` |
| `COLLECTOR_TOKEN` | No | Enables ingestion and allows starting the Codex and Claude logins over HTTP. At least 32 characters and different from `DEVICE_TOKEN` | Output of `openssl rand -hex 32` |
| `AUTH_USERNAME` | No | [Web panel](panel.md) user. Set together with `AUTH_PASSWORD_HASH` | `admin` |
| `AUTH_PASSWORD_HASH` | No | Hash of the panel password. Generated with `make panel-password` | `pbkdf2-sha256:600000:...` |
| `MASTER_KEY` | No | Encrypts the provider credentials saved from the panel. At least 32 characters | Output of `openssl rand -base64 32` |

A plain-text `AUTH_PASSWORD` is not accepted: if it is set, the server does not start.

### Server

| Variable | Default | Purpose |
|---|---|---|
| `CONFIG_NAME` | `config.json` | Name of the accounts file inside `server/configs/`. A name, not a path |
| `HTTP_PORT` | `8080` | Port. In Compose it is the port published on the host |
| `HTTP_HOST` | `127.0.0.1` | Listen address of the binary. Compose sets it to `0.0.0.0` inside the container |
| `HOST_BIND` | `127.0.0.1` | Compose only: host interface where the port is published |
| `POLL_INTERVAL` | `5m` | How often the providers are polled. Go duration, minimum `1s` |
| `STALE_AFTER` | `15m` | Age after which an account becomes `stale`. Must be greater than or equal to `POLL_INTERVAL` |
| `DATA_DIR` | `./data` | Data folder. Compose sets it to `/data`, on the `usage-data` volume |
| `MOCK_FILE` | `configs/mock-usage.json` | Data of the `mock` accounts |
| `CODEX_BIN` | `codex` | Codex CLI executable. The Docker image already includes it |

### Provider credentials

The name of these variables is defined by each account's `credentialEnv` field. The ones in `config.live.json` are:

| Variable | Purpose |
|---|---|
| `OPENROUTER_PERSONAL_KEY` | OpenRouter management key |
| `CURSOR_WORK_SESSION` | Value of the `WorkosCursorSessionToken` cookie from cursor.com |

They are optional if you load the credential from the panel; what is saved in the panel takes priority. Codex and Claude use no variables: they sign in over HTTP or from the panel.

With Docker Compose, `compose.yaml` passes a fixed list of variables to the container. If you add an account with a new `credentialEnv`, add that variable to the `environment` section of `compose.yaml` or load the credential from the panel.

### Optional collector

Only for `server/cmd/collector`, which sends the quotas of a Codex account with `source: collector`:

| Variable | Default | Purpose |
|---|---|---|
| `QUOTA_API_URL` | `http://localhost:8080` | URL of the API that receives the data |
| `CODEX_POLL_INTERVAL` | `5m` | Send interval, minimum `10s` |

### Rules of the `.env` file

- Variables exported in the terminal or set by Docker take priority over `.env`. If you exported `CONFIG_NAME` earlier, run `unset CONFIG_NAME`.
- It accepts simple assignments, quotes and full-line comments.
- It does not run commands or expand variables: paste real values, not shell expressions.

## API tokens

There are three ways to authenticate, independent of each other:

| Credential | Who uses it | What it allows |
|---|---|---|
| `DEVICE_TOKEN` | The display and any read-only client | `GET /usage` |
| `COLLECTOR_TOKEN` | Collectors and command-line administration | `GET /usage`, ingestion and the Codex and Claude logins |
| Panel session | The browser, after signing in | `GET /usage`, the Codex and Claude logins and loading credentials |

All of them are sent as `Authorization: Bearer <token>`. The full per-route table is in [web panel](panel.md#session) and the endpoints in the [API reference](api.md).

## Accounts

The accounts file is a JSON document with an `accounts` list. The repo ships two:

- `server/configs/config.json`: six fictitious accounts (`source: mock`). The default.
- `server/configs/config.live.json`: real OpenRouter, Cursor, personal and work Codex, and Claude accounts.

For your own configuration without versioning it, save it as `server/configs/config.local.json` (it is in `.gitignore`) and set `CONFIG_NAME=config.local.json`. It works the same with Go and with Docker.

```json
{
  "accounts": [
    {"id": "openrouter-personal", "name": "OpenRouter", "provider": "openrouter", "scope": "personal", "source": "direct", "credentialEnv": "OPENROUTER_PERSONAL_KEY"},
    {"id": "codex-work", "name": "Codex work", "provider": "openai", "scope": "work", "source": "direct"}
  ]
}
```

| Field | Values | Notes |
|---|---|---|
| `id` | Lowercase letters, digits, `-` and `_`; up to 80 characters | Unique. It appears in URLs and in session folders |
| `name` | Text, up to 120 characters | Display name |
| `provider` | `openai`, `openrouter`, `cursor`, `anthropic`, `kiro` | `openai` is Codex; `anthropic` is Claude |
| `scope` | `personal`, `work` | Splits the two pages of the panel and the display |
| `source` | `mock`, `direct`, `collector` | Where the data comes from |
| `credentialEnv` | Name of an environment variable | Required for `openrouter` and `cursor` with `source: direct` |

Rules:

- Between 1 and 64 accounts. The JSON is strict: an unknown field prevents startup.
- `source: direct` exists for `openai`, `openrouter`, `cursor` and `anthropic`. `kiro` only accepts `mock` or `collector`.
- Each `mock` account needs an entry with its `id` in `server/configs/mock-usage.json`.
- A `collector` account is not polled: it waits for data through [ingestion](api.md#ingestion).
- Data is stored per `source` and `id` combination: switching an account from `mock` to `direct` does not reuse the fictitious values.

### Personal and work

To have the same service in both scopes, declare it twice with different `id` and `scope`, like `codex-personal` and `codex-work` in `config.live.json`. Each account keeps its own session or credential. The `id` does not check which account you chose when signing in on the browser: use different profiles or a private window to avoid mixing them.

How to authorize each provider: [step-by-step guide](providers/connect.md).
