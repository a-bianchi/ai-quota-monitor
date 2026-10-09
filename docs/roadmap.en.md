# Roadmap

**English** · [Español](roadmap.md)

## Implemented

- Go server without external dependencies.
- GET /usage with personal/work filter and Bearer token.
- Example mode with six accounts.
- OpenRouter API client with errors and timeout.
- Ingestion of observations for collector accounts.
- Persistence of the latest observation and stale-data detection.
- Docker Compose, documentation and Go CI.
- Codex on the server: HTTP device-code login, isolated profiles and automatic polling.

## Next

1. Validate OpenRouter with your own management key.
2. Validate real login and quotas of both accounts with the implemented HTTP login.
3. Implement the firmware and validate LCD/Wi-Fi/TLS on the real board.
4. Validate Cursor with a real session; evaluate non-admin access in Kiro.
5. Validate Claude with a real account; replace it if Anthropic publishes an official path.
6. Design, print and verify the enclosure.

## Open decisions

- Firmware toolchain and UI library.
- Per-quota freshness if there are independent observations.
- Individual tokens per collector and per-account permissions.
- Parametric enclosure model (OpenSCAD or equivalent CAD).
- Web dashboard to preview the screen.

The backend has not had a real deployment or tests with real accounts yet. There are no validated physical assembly instructions.
