# How to connect each provider

**English** · [Español](connect.es.md)

Step-by-step guide. Each provider is connected only once; after that the server queries on its own every `POLL_INTERVAL` (5 minutes by default).

| Provider | How it connects | Where the credential lives | Status |
|---|---|---|---|
| [OpenRouter](#openrouter) | Paste a management key | `server/.env` | Implemented |
| [Codex (ChatGPT)](#codex-chatgpt) | Code login from the browser | `DATA_DIR/codex/<account>/` | Implemented; real login depends on each account |
| [Cursor](#cursor) | Paste the session cookie | `server/.env` | Experimental |
| [Claude](#claude-subscription) | Login from the browser and paste a code | `DATA_DIR/claude/<account>/` | Experimental; not permitted by Anthropic's terms |
| Kiro | — | — | Pending |

## Before you start

1. Copy the example file: `cp server/.env.example server/.env`.
2. Generate two different tokens with `openssl rand -hex 32` and paste them into `DEVICE_TOKEN` and `COLLECTOR_TOKEN`. They belong to this API, not to the providers.
3. Set `CONFIG_NAME=config.live.json` to use real accounts instead of the example ones.
4. Start with `make run` (local) or `make docker-up` (Docker).

Each account is declared in `server/configs/config.live.json`. If you do not use a provider, delete its entry. After changing `.env` or the configuration you have to restart the server.

## How to verify any account

```sh
curl -H 'Authorization: Bearer YOUR_COLLECTOR_TOKEN' http://localhost:8080/usage
```

Look at the account's `status`:

| `status` | What it means | What to do |
|---|---|---|
| `ok` | Recent data | Nothing |
| `stale` | There is data, but it is old or the last query failed | Wait one cycle; if it persists, check the logs |
| `auth_required` | The credential is missing, wrong or expired | Repeat the connection steps for that provider |
| `unavailable` | No data was ever obtained | Check the logs and the server's internet connection |

The first read may take a few seconds after startup.

## OpenRouter

Shows credits purchased and spent, in USD.

1. Go to <https://openrouter.ai/settings/management-keys> and create a **management key**. A regular inference key cannot query credits.
2. Paste it in `server/.env`:

   ```dotenv
   OPENROUTER_PERSONAL_KEY=your_management_key
   ```

3. The account in the configuration:

   ```json
   {"id":"openrouter-personal","name":"OpenRouter","provider":"openrouter","scope":"personal","source":"direct","credentialEnv":"OPENROUTER_PERSONAL_KEY"}
   ```

4. Restart and verify. Expected: a `balance` quota in `usd`.

The key does not expire by itself. If you revoke it, the account goes to `auth_required`.

## Codex (ChatGPT)

Shows the Codex limits of your ChatGPT subscription. They are not all the limits of the regular chat. Requires Codex CLI (the Docker image already includes it; locally it is set with `CODEX_BIN`).

No credential is pasted: you sign in from your browser and the server stores the session.

1. The account in the configuration (the provider is called `openai`):

   ```json
   {"id":"codex-personal","name":"Codex personal","provider":"openai","scope":"personal","source":"direct"}
   ```

2. Request a login code:

   ```sh
   curl -X POST http://localhost:8080/accounts/codex-personal/codex/login \
     -H 'Authorization: Bearer YOUR_COLLECTOR_TOKEN'
   ```

   `202` response with `verificationUrl` and `userCode`.

3. Open `verificationUrl`, sign in with the right account, type the `userCode` and authorize.
4. Check the status with a `GET` on the same endpoint until it says `authenticated`.
5. Verify with `GET /usage`.

To have both personal and work, repeat the steps with each ID (`codex-personal`, `codex-work`). The ID does not check which account you chose in the browser: use different profiles or a private window so they do not get mixed up.

If your organization restricts device-code login, step 3 may fail; that is up to the administrator. More detail in [codex.md](codex.md).

## Cursor

Shows the percentage of the plan used and the reset date. It uses the cursor.com dashboard endpoint, which Cursor does not document and which may change. It does not require being an administrator.

1. Open <https://cursor.com/dashboard> and sign in with the account you want to measure.
2. Open the developer tools (`F12` or `Cmd+Option+I`).
3. Go to **Application** (in Firefox, **Storage**) → **Cookies** → `https://cursor.com`.
4. Find `WorkosCursorSessionToken` and copy the **full value**.
5. Paste it in `server/.env`:

   ```dotenv
   CURSOR_WORK_SESSION=copied_value
   ```

6. The account in the configuration:

   ```json
   {"id":"cursor-work","name":"Cursor trabajo","provider":"cursor","scope":"work","source":"direct","credentialEnv":"CURSOR_WORK_SESSION"}
   ```

7. Restart and verify. Expected: an `included` quota in `percent`, and depending on the plan also `auto`, `api`, `on-demand` or `team-on-demand`.

Keep in mind:

- The cookie gives access to your Cursor account: treat it like a password.
- It expires and does not renew itself. When the account goes to `auth_required`, repeat steps 1 to 5 and restart.
- Signing out of cursor.com may invalidate it earlier.
- Some Enterprise accounts expose no meter with a limit; in that case it stays `unavailable`.

For a second account, add another entry with another variable (for example `CURSOR_PERSONAL_SESSION`) and, if you use Docker, add it to `compose.yaml` as well. More detail in [cursor.md](cursor.md).

## Claude (subscription)

Shows the 5-hour and 7-day windows of Claude Pro/Max.

**Before using it:** Anthropic's terms do not permit a third-party application to sign in with Claude.ai or store its tokens, and that is what this provider does. The risk is on your account. Read the [full warning](claude-subscription.md).

1. The account in the configuration:

   ```json
   {"id":"claude-personal","name":"Claude personal","provider":"anthropic","scope":"personal","source":"direct"}
   ```

2. Request the login URL:

   ```sh
   curl -X POST http://localhost:8080/accounts/claude-personal/claude/login \
     -H 'Authorization: Bearer YOUR_COLLECTOR_TOKEN'
   ```

3. Open `verificationUrl`, sign in with the right account and authorize. Copy the code the page shows (`code#state`).
4. Send it within 15 minutes:

   ```sh
   curl -X PUT http://localhost:8080/accounts/claude-personal/claude/login \
     -H 'Authorization: Bearer YOUR_COLLECTOR_TOKEN' \
     -H 'Content-Type: application/json' \
     -d '{"code":"code#state"}'
   ```

5. Verify with `GET /usage`. Expected: `five-hour` and `seven-day` quotas in `percent`.

The server renews the session by itself. If it goes to `auth_required`, repeat steps 2 to 4. More detail in [claude-subscription.md](claude-subscription.md).

## Security

- `server/.env` and `server/data/` are in `.gitignore`. Do not commit them or paste them in issues.
- In Docker, the Codex and Claude sessions live in the `usage-data` volume; its backups are secrets too.
- Use `DEVICE_TOKEN` on the display and `COLLECTOR_TOKEN` only for administration.
