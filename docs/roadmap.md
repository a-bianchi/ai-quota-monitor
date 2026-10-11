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
- ESP32-C6 firmware with Arduino IDE and LVGL 8.3: it has a step-by-step guide and runs on the physical board.
- Two-part 3D-printable enclosure with a snap-fit lid: parametric model, STL and 3MF files; printed.
- Deployment on a local Coolify server.

## Next

1. Validate OpenRouter with your own management key.
2. Validate real login and quotas of both accounts with the implemented HTTP login.
3. Test the firmware over HTTPS with a real certificate, Wi-Fi reconnection and long runs.
4. Validate Cursor with a real session; implement Kiro if it is worth it (see [providers](providers/README.md)).
5. Validate Claude with a real account; replace it if Anthropic publishes an official path.
6. Document the material and print profile of the enclosure.

## Open decisions

- Per-quota freshness if there are independent observations.
- Individual tokens per collector and per-account permissions.

The experimental integrations have not been validated with real accounts yet.
