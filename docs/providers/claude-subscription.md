# Claude (Pro/Max subscription)

**English** · [Español](claude-subscription.es.md)

Shows the limits of the Claude subscription: 5-hour window, 7-day window and, when present, the per-model weekly ones. These are not Anthropic API credits.

## Warning: read before using

Status: **experimental and not permitted by Anthropic's terms.**

Anthropic publishes no API to read the limits of a subscription. This provider signs in with the public OAuth client of Claude Code and stores the tokens on the server. The [Claude Code legal page](https://code.claude.com/docs/en/legal-and-compliance) says that third-party developers may not offer Claude.ai login in their applications, nor collect or store Claude.ai credentials or session tokens, and that Anthropic may enforce those restrictions without notice.

Using it is your decision and the risk is on your account. If you do not want to take it, remove the `anthropic` account from your configuration: the rest of the monitor works the same.

In addition, the usage endpoint is not documented and may change or stop working. The integration was tested only with simulated endpoints.

## What it queries

- Login: OAuth authorization with PKCE, requesting only the `user:profile` permission. The session cannot be used to send prompts.
- Usage: `GET https://api.anthropic.com/api/oauth/usage`. It sends no prompts and consumes no quota.

## Configuration

In `server/configs/config.live.json`:

```json
{"id":"claude-personal","name":"Claude personal","provider":"anthropic","scope":"personal","source":"direct"}
```

Requires `COLLECTOR_TOKEN`. Restart the server after changing the configuration.

## HTTP login

1. Start:

   ```sh
   curl -X POST http://YOUR_SERVER:8080/accounts/claude-personal/claude/login \
     -H 'Authorization: Bearer YOUR_COLLECTOR_TOKEN'
   ```

   `202` response:

   ```json
   {"status":"pending","verificationUrl":"https://claude.com/cai/oauth/authorize?...","expiresAt":"2026-10-09T16:00:00Z"}
   ```

2. Open `verificationUrl` in your browser, sign in with the right account and authorize. The page shows a code shaped like `code#state`.

3. Send the full code:

   ```sh
   curl -X PUT http://YOUR_SERVER:8080/accounts/claude-personal/claude/login \
     -H 'Authorization: Bearer YOUR_COLLECTOR_TOKEN' \
     -H 'Content-Type: application/json' \
     -d '{"code":"code#state"}'
   ```

   `200` response with `{"status":"authenticated"}`. The server queries the limits right away.

4. Status, whenever you want: `GET /accounts/claude-personal/claude/login`.

Errors in step 3: `409` if there is no login in progress or the 15 minutes passed, `400` if the code was rejected, `502` if the communication with Claude failed. In all three cases start again from step 1.

Statuses: `pending`, `authenticated`, `auth_required`, `unknown`. After a restart the status is `unknown` until the first read; it does not mean the session was lost.

## Returned quotas

| ID | Meaning |
|---|---|
| `five-hour` | 5-hour window |
| `seven-day` | 7-day window, all models |
| `seven-day-opus`, `seven-day-sonnet` | Per-model weekly windows, if the response includes them |

All in `percent`, with `resetsAt` when Claude reports it.

## Persistence and limits

- The session is stored in `DATA_DIR/claude/<account-id>/credentials.json`; in Docker, inside the `usage-data` volume. It is a secret: do not expose it or commit it to Git.
- The server renews the session by itself. If the renewal is rejected, the account goes to `auth_required` and the login has to be repeated.
- The account ID does not verify which account you signed in with in the browser.
- It was not confirmed with a real account that the `user:profile` permission is enough to read the usage.
