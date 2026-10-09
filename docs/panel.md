# Web panel

**English** · [Español](panel.es.md)

A small web panel served by the same Go API at `/`. It is plain HTML, CSS and Alpine.js embedded in the server binary: no Node, no build step and no separate process. It works the same with `make run` and with Docker Compose.

## Views

- **Services**: personal and work accounts, with provider, status, quotas and last update.
- **Display**: preview of the 172 × 320 px ESP32 screen with the real API data, switching between personal and work.
- **Settings**: Codex login (code and authorization link shown in the panel), Claude login, and loading or replacing the OpenRouter key and the Cursor cookie.

Spanish and English, with a selector; the choice is saved in the browser. Spanish is the initial language.

## Create the user

There are no default credentials. The panel stays disabled until you define a user.

1. Generate the password hash. It asks for the password twice and does not show it:

   ```sh
   make panel-password
   ```

   It prints one line, `AUTH_PASSWORD_HASH=pbkdf2-sha256:...`. The password needs at least 6 characters.

2. Paste that line into `server/.env` and choose the username:

   ```dotenv
   AUTH_USERNAME=your_username
   AUTH_PASSWORD_HASH=pbkdf2-sha256:600000:...
   ```

3. Optional, to save provider credentials from the panel, add a master key:

   ```sh
   openssl rand -base64 32
   ```

   ```dotenv
   MASTER_KEY=the_generated_value
   ```

4. Restart (`make run` or `make docker-up`) and open <http://localhost:8080/>.

With Docker and no local Go, generate the hash inside the image:

```sh
docker compose --env-file server/.env run --rm -it api hash-password
```

To change the password, generate a new hash, replace the line and restart. `AUTH_PASSWORD` in plain text is no longer accepted: the server refuses to start if it is set.

## Session

- `POST /auth/login` returns an access token (15 minutes) and a refresh token (30 days without use). The panel keeps both in `localStorage` and sends `Authorization: Bearer <accessToken>`. No cookies are used.
- The panel renews the access token shortly before it expires. If a request gets a 401 it renews and repeats it once; if the renewal fails it clears the session and returns to the login.
- Every renewal rotates the refresh token and invalidates the previous one. Using an already rotated refresh token revokes that session.
- The server stores only SHA-256 hashes of the tokens, in `DATA_DIR/panel-sessions.json`, so sessions survive restarts. Up to 32 sessions are kept.
- `POST /auth/logout` revokes the session on the server and the panel deletes the tokens from the browser.
- Login attempts are limited: 5 failures per client address and 50 in total every 15 minutes; after that the server answers 429.
- The password is stored only as a PBKDF2-HMAC-SHA256 hash with 600,000 iterations and a random salt.

Panel tokens are independent of `DEVICE_TOKEN` and `COLLECTOR_TOKEN`, which keep working as before for the ESP32 and collectors. The panel never stores or needs those two tokens.

| Route | Panel session | `DEVICE_TOKEN` | `COLLECTOR_TOKEN` |
|---|---|---|---|
| `GET /usage` | yes | yes | yes |
| `GET /accounts`, `PUT /accounts/{id}/credential` | yes | no | no |
| Codex and Claude login routes | yes | no | yes |
| `PUT /accounts/{id}/usage` (ingestion) | no | no | yes |

The HTML, CSS and JavaScript files are public, like any login page; they contain no data. Every route that returns or changes data checks the session on the server.

## Provider credentials

- Saved on the server in `DATA_DIR/credentials.json`, encrypted with AES-256-GCM using a key derived from `MASTER_KEY`.
- Never returned to the browser: the panel only shows whether one is configured and where it comes from.
- **Priority**: the value saved from the panel wins; if there is none, the environment variable named in `credentialEnv` (for example `OPENROUTER_PERSONAL_KEY`) is used as a fallback.
- Without `MASTER_KEY` the panel cannot save credentials; environment variables keep working.
- If `MASTER_KEY` changes or is lost, the saved credentials cannot be read and have to be loaded again.
- Codex and Claude keep their own per-account sessions on the server; they are not API keys.

## Security notes

- `localStorage` can be read by JavaScript running on the page, so the panel loads no third-party scripts, never inserts HTML from data, and is served with a strict Content Security Policy (`script-src 'self'`, no inline code, no `eval`). Alpine.js is vendored at a fixed version in its CSP build; see `server/internal/web/static/vendor/README.md`.
- Use HTTPS when the panel is reached from outside a trusted network: without it, the password and tokens travel in clear. See [deployment](deployment.md).
- Behind a reverse proxy all clients share one address, so the per-address login limit behaves like a global one.
- Sessions, credentials and provider sessions live in the data volume: treat it and its backups as secrets.

## What is verified

Automated tests cover login, rotation, reuse detection, logout, expiry, the attempt limit, route protection, credential encryption and priority, and the static files. The interface was exercised end to end against a running server with example accounts. Not verified with real accounts from the panel: a complete Codex or Claude login and reading OpenRouter with a key saved from the panel.
