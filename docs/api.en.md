# API v1

**English** · [Español](api.md)

## Login with username and password

Optional. Enabled by setting `AUTH_USERNAME` and `AUTH_PASSWORD` (8+ characters) together with `COLLECTOR_TOKEN`; without them, these endpoints answer 503. `DEVICE_TOKEN` and `COLLECTOR_TOKEN` keep working the same way.

`POST /auth/login`, without `Authorization`:

```json
{"username": "your_username", "password": "your_password"}
```

200 response:

```json
{"tokenType": "Bearer", "accessToken": "<jwt>", "expiresIn": 900, "refreshToken": "<jwt>"}
```

`POST /auth/refresh` with `{"refreshToken": "<jwt>"}` returns the same shape, with new tokens.

- `accessToken` lasts 15 minutes and is used as `Authorization: Bearer <accessToken>`. It grants the same permissions as `COLLECTOR_TOKEN`: reading `/usage`, ingesting data and starting provider logins.
- `refreshToken` lasts 30 days and only works for `/auth/refresh`.
- 401 if the credentials or the refresh token are invalid or expired; 400 if the JSON is invalid.

The tokens are signed JWTs (HS256) and the server does not store them: they remain valid after a restart. There is no logout or per-token revocation; to invalidate all of them, change `AUTH_PASSWORD` or `COLLECTOR_TOKEN`. A failed login takes one second to answer; there is no lockout after repeated attempts, so use a strong password and HTTPS if the API leaves your network.

## Reading

`GET /usage`, with `Authorization: Bearer DEVICE_TOKEN`. It also accepts `COLLECTOR_TOKEN`, so everything can be tested with a single token from Postman. Optional filter `?scope=personal` or `?scope=work`. `GET /healthz` is public and only indicates that the HTTP process responds; it does not guarantee that the providers are available.

Optional headers:

| Header | Values | Without the header |
|---|---|---|
| `X-Timezone` | IANA zone, for example `Europe/Madrid` or `UTC` | `America/Argentina/Buenos_Aires` (UTC-3) |
| `Accept-Language` | `es` or `en` | `es` |

`X-Timezone` converts `generatedAt`, `updatedAt` and `resetsAt` to that zone; the instant is the same, only the offset changes (`-03:00`, `+02:00`, or `Z` for UTC). An unknown zone returns 400. `Accept-Language` translates `quotas[].name`; the first supported language in the list is used and anything else falls back to Spanish. Quota `id` values are never translated. Account names come from the configuration and are not translated, and neither are custom names sent by a collector.

Response (with `Accept-Language: en`):

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
      "resetsAt": "2026-10-08T20:20:00-03:00"
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
- `used`, `limit`, `updatedAt`, `resetsAt` may be `null` where applicable.
- `quotas` is a list, empty while there is no data yet.

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

Only accounts declared with `source: collector` are accepted; identity metadata comes from the configuration, not from the payload. Maximum 64 KiB and 16 quotas, unique IDs and finite non-negative values. It rejects observations older than the latest data and future timestamps beyond a one-minute margin. The list replaces all quotas of the account. Response 200 when persisted, 400 invalid payload, 401 wrong token, 404 account not ingestible, 409 older observation and 503 if ingestion is disabled.

`COLLECTOR_TOKEN` is shared by the collectors of this MVP: it can update every collector account. It also allows reading `GET /usage`. Do not use it as the device token. Do not send provider credentials in the report.

## Codex login on the homelab

POST and GET `/accounts/{id}/codex/login`, with `Authorization: Bearer COLLECTOR_TOKEN`. Only accounts with provider=openai, source=direct. POST starts the device-code login and returns the official code/URL; GET reports the status. DEVICE_TOKEN has no access. If COLLECTOR_TOKEN is empty, the endpoints are disabled (503). See the [flow](providers/codex.en.md).

## Claude login on the homelab

`/accounts/{id}/claude/login`, with `Authorization: Bearer COLLECTOR_TOKEN`. Only accounts with provider=anthropic, source=direct. POST starts the login and returns `verificationUrl`; PUT with `{"code":"code#state"}` completes it; GET reports the status. DEVICE_TOKEN has no access. Experimental and not permitted by Anthropic's terms: see the [flow and warning](providers/claude-subscription.en.md).
