# Roadmap

**English** · [Español](roadmap.es.md)

## Implemented

- Go server without external dependencies.
- GET /usage with personal/work filter and Bearer token.
- Example mode with six accounts.
- OpenRouter API client with errors and timeout.
- Ingestion of observations for collector accounts.
- Persistence of the latest observation and stale-data detection.
- Docker Compose, documentation and Go CI.
- Web panel embedded in the server: login with rotating sessions, dashboard, ESP32 screen preview and provider configuration.
- Codex on the server: HTTP device-code login, isolated profiles and automatic polling.
- Experimental clients for Cursor and the Claude subscription.
- Provider credentials loaded from the panel and stored encrypted.
- Per-quota `usageLevel` in `GET /usage`.
- Quota names in Spanish or English and dates in the requested time zone.
- ESP32-C6 firmware with Arduino IDE and LVGL 8.3: it compiles and has a step-by-step guide; not tested on the board.

## Next

1. Validate OpenRouter with your own management key.
2. Validate real login and quotas of both accounts with the implemented HTTP login.
3. Validate the firmware on the real board: LCD, Wi-Fi and TLS.
4. Validate Cursor with a real session; evaluate non-admin access in Kiro.
5. Validate Claude with a real account; replace it if Anthropic publishes an official path.
6. Design, print and verify the enclosure.

## Open decisions

- Per-quota freshness if there are independent observations.
- Individual tokens per collector and per-account permissions.
- Parametric enclosure model (OpenSCAD or equivalent CAD).

The backend has not had a real deployment or tests with real accounts yet. There are no validated physical assembly instructions.
