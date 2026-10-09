# Codex on the server / homelab

**English** · [Español](codex.es.md)

The Go API starts the device-code login through Codex App Server. The user authorizes in their browser and Codex stores and renews the credentials on the server. No collector on your Mac is required, and no cookies or access tokens are copied into the API.

## Configuration

In `server/configs/config.live.json`, each Codex account uses:

```json
{"id":"codex-work","name":"Codex trabajo","provider":"openai","scope":"work","source":"direct"}
```

In `server/.env`:

```dotenv
CONFIG_NAME=config.live.json
COLLECTOR_TOKEN=a_random_token_of_at_least_32_characters
```

Generate the token with `openssl rand -hex 32`. It must be different from DEVICE_TOKEN. Reuse the existing COLLECTOR_TOKEN: it is shared by all providers and allows ingesting data and starting/checking logins; it is not an OpenAI credential. Do not use the literal placeholder. Restart the server after changing configuration or tokens.

- Local Go: `make run`; requires Codex CLI, configurable with CODEX_BIN.
- Homelab/Docker: `make docker-up`; the image includes Codex 0.162.0. Compose uses CODEX_BIN=codex, without the path of the executable on your Mac.
- LAN: adjust HOST_BIND as described in [deployment](../deployment.md).

## HTTP login

```sh
curl -X POST http://YOUR_SERVER:8080/accounts/codex-work/codex/login \
  -H 'Authorization: Bearer YOUR_COLLECTOR_TOKEN'
```

Example 202 response:

```json
{
  "status":"pending",
  "verificationUrl":"https://auth.openai.com/codex/device",
  "userCode":"ABCD-1234",
  "expiresAt":"2026-10-09T16:00:00Z"
}
```

Open verificationUrl in your browser, enter the code and choose the work account/workspace. For the personal one use the codex-personal endpoint. It is your regular access; your organization's restrictions still apply. You may need to enable device-code login in the ChatGPT security settings, or ask your administrator to enable the method if it is restricted.

The example code is not valid. expiresAt is the local wait limit (15 minutes); OpenAI may invalidate the code earlier. Do not share userCode or the login response in an issue.

Status:

```sh
curl http://YOUR_SERVER:8080/accounts/codex-work/codex/login \
  -H 'Authorization: Bearer YOUR_COLLECTOR_TOKEN'
```

Statuses: starting, pending, authenticated, auth_required, failed, unknown. After a restart, the status is recovered when the session is queried. unknown does not mean the session was lost. A repeated POST while it is pending returns the same attempt, without launching another login. If a read is in progress it may answer 409; retry later.

When the login completes, the server queries the quotas immediately. It then refreshes every POLL_INTERVAL (5m by default). GET /usage accepts DEVICE_TOKEN or COLLECTOR_TOKEN; DEVICE_TOKEN cannot start logins.

## Persistence and limits

Sessions live in DATA_DIR/codex/<account-id>/; in Docker that is /data/codex/<account-id>/ inside the usage-data volume. Each account has separate credentials. Restarting keeps the sessions; deleting the volume removes them. Treat the volume and its backups as secrets. Do not expose this folder or put it in Git.

Only account endpoints are used: initialize, account/login/start (chatgptDeviceCode), account/read and account/rateLimits/read. There are no prompts, tools or use of reset credits. The API neither receives nor returns cookies or access/refresh tokens. Codex is started with plugins disabled, since they are not needed to read limits.

The ID codex-work does not verify which account you chose: check the account/workspace during login. The distinction between real buckets and their durations is preserved; no windows are invented that do not appear in the response. These limits belong to Codex/Work, not to every ChatGPT chat.

For personal self-hosted use. OpenAI's documentation distinguishes local/open-source integrations from commercial/hosted services with App Server authentication; do not turn these endpoints into a public multi-user service. Check the current requirements before offering it as a service.

[Official App Server](https://learn.chatgpt.com/docs/app-server) · [Authentication](https://learn.chatgpt.com/docs/auth)

The flow is verified with a simulated App Server; real logins and policies depend on each account. Docker must be able to reach OpenAI over HTTPS.

## Earlier local collector

The cmd/collector binary remains as an option for source=collector configurations. It is not needed for config.live.json or for the current homelab deployment.
