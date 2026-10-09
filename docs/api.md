# API v1

**English** · [Español](api.es.md)

A Postman collection and environment for these endpoints live in [`docs/collections/`](collections/). Import both files, select the `ai-quota-monitor` environment and fill in `url` (for example `http://localhost:8080`), `COLLECTOR_TOKEN` and, if you enabled the panel, `username` and `password`. The login and refresh requests save the tokens in the environment. Do not commit your values.

The API answers JSON. The examples use `http://localhost:8080`; replace the `YOUR_...` placeholders with your values.

## Endpoint summary

| Method and route | Authentication | Use |
|---|---|---|
| `GET /ping` | None | Check that the process responds |
| `GET /usage` | `DEVICE_TOKEN`, `COLLECTOR_TOKEN` or panel session | [Read the quotas](#reading) |
| `PUT /accounts/{id}/usage` | `COLLECTOR_TOKEN` | [Send data from a collector](#ingestion) |
| `POST`, `GET /accounts/{id}/codex/login` | `COLLECTOR_TOKEN` or panel session | [Codex login](#codex-login-on-the-homelab) |
| `POST`, `PUT`, `GET /accounts/{id}/claude/login` | `COLLECTOR_TOKEN` or panel session | [Claude login](#claude-login-on-the-homelab) |
| `POST /auth/login`, `/auth/refresh`, `/auth/logout` | Username and password, or refresh token | [Panel session](#panel-session) |
| `GET /accounts` | Panel session | [Configurable accounts](#panel-configuration) |
| `PUT /accounts/{id}/credential` | Panel session | [Save a credential](#panel-configuration) |
| `GET /` and `/static/` | None | [Web panel](panel.md) files |

Tokens are sent as `Authorization: Bearer <token>`. What each one allows: [API tokens](configuration.md#api-tokens).

Errors always have the same shape, with a non-2xx HTTP code:

```json
{"error": "unauthorized"}
```

Codes common to protected routes: 401 if the token is missing or wrong, and 503 (`endpoint disabled`) if the route depends on something that is not configured.

## Panel session

Used by the [web panel](panel.md); optional. Enabled by setting `AUTH_USERNAME` and `AUTH_PASSWORD_HASH`; without them these endpoints answer 503. `DEVICE_TOKEN` and `COLLECTOR_TOKEN` are independent and keep working the same way.

`POST /auth/login`, without `Authorization`:

```json
{"username": "your_username", "password": "your_password"}
```

200 response:

```json
{"tokenType": "Bearer", "accessToken": "aqm_a_...", "expiresIn": 900, "refreshToken": "aqm_r_..."}
```

- `POST /auth/refresh` with `{"refreshToken": "..."}` returns the same shape with new tokens. The refresh token used stops working; presenting it again revokes the session.
- `POST /auth/logout` revokes the session of the access token in `Authorization` or of `{"refreshToken": "..."}` in the body. It answers 204.
- `accessToken` lasts 15 minutes; `refreshToken` lasts 30 days without use and only works for `/auth/refresh` and `/auth/logout`.
- 401 for wrong credentials or an invalid, expired or already used refresh token; 400 for invalid JSON; 429 with `Retry-After` after too many failed logins.

The server stores only hashes of the tokens, so sessions survive restarts. See [which routes accept a panel session](panel.md#session).

## Panel configuration

Panel session only; the static tokens are not accepted.

- `GET /accounts` lists the configured accounts and what can be set up for each: `login` (`codex`, `claude` or empty) and `credential` (`null`, or `{"configured": true, "origin": "stored" | "env"}`). `credentialStorage` tells whether `MASTER_KEY` is set. Credential values are never returned.
- `PUT /accounts/{id}/credential` with `{"value": "..."}` stores or replaces the credential of an account that uses `credentialEnv`. 404 if the account does not use one, 400 for an empty or multi-line value, 503 without `MASTER_KEY`.

## Reading

`GET /usage`, with `Authorization: Bearer DEVICE_TOKEN`. It also accepts `COLLECTOR_TOKEN` and a panel session. Optional filter `?scope=personal` or `?scope=work`. `GET /ping` is public and only indicates that the HTTP process responds; it does not guarantee that the providers are available.

Optional headers:

| Header | Values | Without the header |
|---|---|---|
| `X-Timezone` | IANA zone, for example `Europe/Madrid` or `UTC` | `America/Argentina/Buenos_Aires` (UTC-3) |
| `Accept-Language` | `es` or `en` | `es` |

`X-Timezone` converts `generatedAt`, `updatedAt` and `resetsAt` to that zone; the instant is the same, only the offset changes (`-03:00`, `+02:00`, or `Z` for UTC). An unknown zone returns 400. `Accept-Language` translates `quotas[].name`; the first supported language in the list is used and anything else falls back to Spanish. Quota `id` values are never translated. Account names come from the configuration and are not translated, and neither are custom names sent by a collector.

Examples:

```sh
curl http://localhost:8080/ping

curl -H 'Authorization: Bearer YOUR_DEVICE_TOKEN' \
  'http://localhost:8080/usage?scope=work'

curl -H 'Authorization: Bearer YOUR_DEVICE_TOKEN' \
  -H 'X-Timezone: UTC' -H 'Accept-Language: en' \
  http://localhost:8080/usage
```

`GET /ping` answers `{"status":"ok"}`. Response of `GET /usage` (with `Accept-Language: en`):

```json
{
  "schemaVersion": 1,
  "generatedAt": "2026-10-08T17:41:00-03:00",
  "accounts": [{
    "id": "codex-work",
    "name": "Codex trabajo",
    "provider": "openai",
    "scope": "work",
    "source": "collector",
    "updatedAt": "2026-10-08T17:40:00-03:00",
    "status": "ok",
    "quotas": [{
      "id": "5h",
      "name": "5 hours",
      "unit": "percent",
      "used": 11,
      "limit": 100,
      "resetsAt": "2026-10-08T20:20:00-03:00",
      "usageLevel": "low"
    }]
  }]
}
```

The values above are fictitious. Dates are RFC3339/ISO 8601, in Argentina time (UTC-3) unless another zone is requested with `X-Timezone`. `updatedAt` is the provider observation; `generatedAt` is when this response was generated. Do not use `generatedAt` to infer freshness.

- `provider`: `openai`, `cursor`, `kiro`, `openrouter`, `anthropic`.
- `scope`: `personal`, `work`.
- `source`: `mock`, `direct`, `collector`.
- `status`: `ok`, `stale`, `auth_required`, `unavailable`.
- `unit`: `percent`, `credits`, `usd`, `tokens`.
- `usageLevel`: `low`, `medium`, `high`, `unknown`. See [usage level](#usage-level).
- `used`, `limit`, `updatedAt`, `resetsAt` may be `null` where applicable.
- `quotas` is a list, empty while there is no data yet.

### Usage level

`usageLevel` classifies each quota by the percentage used, `used / limit × 100`. The server computes it on every response, for any unit:

| Percentage used | `usageLevel` | Suggested color |
|---|---|---|
| Under 50% | `low` | Green |
| 50% to 80% inclusive | `medium` | Orange |
| Above 80% | `high` | Red |
| `used` or `limit` missing, or a limit of 0 | `unknown` | Gray |

The colors are a suggestion for clients; the API only returns the level. The field appears only in responses: an ingestion report that includes it is rejected with 400. For OpenRouter the percentage compares credits spent with credits purchased, so it does not represent a monthly quota.

### Interpreting the values

Balance = `limit - used` when both exist. If consumption exceeds the limit, show it as exhausted and keep the original values. For percentages the limit is 100. Do not add up percentages of different quotas. OpenRouter returns accumulated credits purchased and used: show the USD balance; that total is neither a monthly quota nor a subscription bar.

`stale`: `STALE_AFTER` elapsed, a window already reset, or an update failed and earlier data exists. `auth_required`: key missing or rejected. `unavailable`: no usable observation or a storage error. Earlier values may still be present with an error: the display must show their status.

The MVP has one timestamp/status per account. A collector must send the quotas of a single observation together; independent data will require extending the contract.

## Ingestion

`PUT /accounts/{id}/usage` with `Authorization: Bearer COLLECTOR_TOKEN`:

```json
{
  "updatedAt": "2026-10-08T20:40:00Z",
  "quotas": [{"id":"5h","name":"5 horas","unit":"percent","used":11,"limit":100,"resetsAt":null}]
}
```

Only accounts declared with `source: collector` are accepted; identity metadata comes from the configuration, not from the payload. Maximum 64 KiB and 16 quotas, unique IDs and finite non-negative values. It rejects observations older than the latest data and future timestamps beyond a one-minute margin. The list replaces all quotas of the account. Response 200 with `{"status":"ok"}` when persisted, 400 invalid payload, 401 wrong token, 404 account not ingestible, 409 older observation, 500 if it could not be stored and 503 if ingestion is disabled. There is a full `curl` example in [local setup](local-setup.md#testing-ingestion-by-hand).

`COLLECTOR_TOKEN` is shared by the collectors of this MVP: it can update every collector account. It also allows reading `GET /usage`. Do not use it as the device token. Do not send provider credentials in the report.

## Codex login on the homelab

POST and GET `/accounts/{id}/codex/login`, with `Authorization: Bearer COLLECTOR_TOKEN` or a panel session. Only accounts with provider=openai, source=direct. POST starts the device-code login and returns the official code/URL; GET reports the status. DEVICE_TOKEN has no access. If neither COLLECTOR_TOKEN nor the panel login is configured, the endpoints are disabled (503). See the [flow](providers/codex.md).

## Claude login on the homelab

`/accounts/{id}/claude/login`, with `Authorization: Bearer COLLECTOR_TOKEN` or a panel session. Only accounts with provider=anthropic, source=direct. POST starts the login and returns `verificationUrl`; PUT with `{"code":"code#state"}` completes it; GET reports the status. DEVICE_TOKEN has no access. Experimental and not permitted by Anthropic's terms: see the [flow and warning](providers/claude-subscription.md).
