# AI Quota Monitor

**English** · [Español](README.es.md)

Desk monitor that shows the remaining balance of your AI services, with separate pages for **Personal** and **Work**, built on a Waveshare ESP32-C6-LCD-1.47.

**Status: server MVP.** Go API, example mode, persistence and OpenRouter client implemented. Codex integrated in the server with device-code login and isolated profiles. Firmware, other collectors and the enclosure are still pending. No data or credentials from real accounts are published.

## Getting started with Docker Compose

Requirements: Git and Docker with Compose v2.

```sh
cp server/.env.example server/.env
openssl rand -hex 32
```

Copy the result into `DEVICE_TOKEN` in `server/.env`. Generate a different token for `COLLECTOR_TOKEN` if you are going to receive metrics from a collector. Then:

```sh
make docker-up
curl http://localhost:8080/healthz
```

To query data, replace the placeholder with your token:

```sh
curl -H 'Authorization: Bearer YOUR_DEVICE_TOKEN' \
  'http://localhost:8080/usage?scope=personal'
```

The initial configuration shows **six fictitious accounts**, with `source: mock`. `used` means consumption; the display must compute the remaining balance. The examples do not prove real access to Cursor, Kiro or Claude. Loading the first batch may take a moment.

## Running without Docker

Requirement: Go 1.22 or later; on macOS, Xcode Command Line Tools for the system linker. From the repo root:

```sh
cp server/.env.example server/.env
# Fill in DEVICE_TOKEN in server/.env with a token generated with openssl rand -hex 32.
make run
```

The server listens on `127.0.0.1:8080`. Data is stored in `server/data/`. The server loads `server/.env` automatically when it runs from `server/` (as `make run` does). Variables exported in the terminal take priority. Docker uses the same file through `make docker-up`.

```sh
make test
make build
```

To use real OpenRouter data, set `CONFIG_NAME=config.live.json` and your management key in `server/.env`. Restart with `make run` or `make docker-up`.

## Integrations

| Service | Repo status | Source |
|---|---|---|
| OpenRouter | Client implemented | Official credits API; requires a management key |
| Codex / Work | Integrated in the server; HTTP device-code login and automatic polling | App Server `account/rateLimits/read` |
| Claude Pro/Max | Integrated in the server, experimental; [not permitted by Anthropic's terms](docs/providers/claude-subscription.md) | OAuth login and undocumented usage endpoint |
| Cursor | Client implemented, experimental; not validated with a real account | Undocumented dashboard endpoint (`/api/usage-summary`) with the session cookie |
| Kiro | Research pending for non-admin access | Credits panel; automatic path still to be validated |
| Regular ChatGPT | Not integrated | No complete counter was confirmed through an API |

To connect your accounts, follow the [step-by-step guide](docs/providers/connect.md). See also [providers](docs/providers/README.md). Codex quotas do not represent every ChatGPT feature.

## Querying Codex from the homelab

With `CONFIG_NAME=config.live.json`, the API queries Codex directly. It needs no process on your Mac. Use your existing `COLLECTOR_TOKEN` (32+ characters and different from DEVICE_TOKEN) in `server/.env`, restart the API and send:

```sh
curl -X POST http://localhost:8080/accounts/codex-work/codex/login \
  -H 'Authorization: Bearer YOUR_COLLECTOR_TOKEN'
```

The response returns `verificationUrl` and `userCode`. Open that URL in your browser, enter the code and choose the work account. The session stays on the server. Check the status with a GET on the same URL and the same token. When the login completes, the quotas are updated automatically.

For your personal account use `codex-personal`. See the [server login guide](docs/providers/codex.md).

## Architecture

```text
OpenRouter + Codex ──> Go server ──> GET /usage ──> ESP32
                            ↑
Browser ── authorizes the Codex device-code login
```

The server queries the direct OpenRouter/Codex sources and keeps the optional ingestion for other accounts configured as `collector`. It stores the latest value and flags stale data or authentication problems. Reads from the ESP32 do not trigger new requests to the provider.

## Postman collection

`docs/collections/` contains a ready-to-use Postman collection and environment:

- `ai-quota-monitor.postman_collection.json`: login, refresh, `GET /usage` and the Codex and Claude login requests.
- `ai-quota-monitor.postman_environment.json`: the variables the requests use, with empty values.

Import both files in Postman (**Import** → drop the two files), select the `ai-quota-monitor` environment and fill in `url` (for example `http://localhost:8080`), `COLLECTOR_TOKEN` and, if you enabled the login, `username` and `password`. The login and refresh requests save `access_token` and `refresh_token` in the environment automatically. Keep your values in Postman; do not commit them.

## Documentation

- [Contract and endpoints](docs/api.md)
- [Postman collection and environment](docs/collections/)
- [Local setup and variables](docs/local-setup.md)
- [Docker, homelab, Coolify and cloud](docs/deployment.md)
- [Hardware and bill of materials](docs/hardware.md)
- [Firmware and screen design](docs/firmware.md)
- [Enclosure and 3D printing](enclosure/README.md)
- [Roadmap](docs/roadmap.md)

JSON configuration lives in `server/configs/`; secrets in `server/.env` or environment variables. This first version uses only the Go standard library. Mock mode lets you develop without hardware.

## Contributing

Run `make test` and `cd server && go vet ./...` before proposing changes. Every integration must document authentication, units, refresh behavior and limitations. Do not include cookies, session files, keys or work information in commits or issues.

MIT license. Not affiliated with the manufacturers or providers mentioned.
