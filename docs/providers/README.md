# Providers

**English** · [Español](README.es.md)

To connect your accounts, see the [step-by-step guide](connect.md).

The specific implementation of each service lives in `server/internal/providers/<provider>/`:

- `codex/`: Codex App Server client, authentication, sessions and quota transformation.
- `openrouter/`: HTTP client and credits query.
- `cursor/`: HTTP client for the dashboard usage summary.
- `claude/`: OAuth login, session renewal and reading of the subscription limits.

The commands and the HTTP API use these packages; future providers must follow the same structure.

## OpenRouter — client implemented

`source: direct`, `provider: openrouter`, `credentialEnv` points to a server variable. It queries `/api/v1/credits` with Bearer and a management key. It returns total purchased and total used, in USD. The balance is the difference; it is not a monthly allowance. 401/403 produce auth_required. No real credentials were tested while the project was created.

[Official documentation](https://openrouter.ai/docs/api/api-reference/credits/get-remaining-credits)

## Codex / Work — integrated in the server

Query your own quotas with `account/rateLimits/read` from Codex App Server, using the corresponding authentication. There may be several buckets; keep the real duration and reset instead of always assuming 5h/weekly. Validate isolated sessions for personal/work accounts without changing the main session. The API starts the code login over HTTP and queries the quotas directly on the homelab. See [Codex setup](codex.md).

[Official documentation](https://learn.chatgpt.com/docs/app-server)

The data does not represent every quota of the regular ChatGPT chat.

## Claude — integrated in the server, experimental

`source: direct`, `provider: anthropic`. OAuth login over HTTP and reading of the 5h and 7d windows from an undocumented endpoint. It is the subscription, not API spend. **Anthropic's terms do not permit this use of the credentials**; read the warning in [Claude setup](claude-subscription.md) before enabling it. Not validated with a real account.

## Cursor — client implemented, experimental

`source: direct`, `provider: cursor`, `credentialEnv` points to the cursor.com session cookie. It queries the undocumented `/api/usage-summary` endpoint with regular user permissions, without the Admin API. It returns percentages with the billing-cycle reset. Not validated with a real account. See [Cursor setup](cursor.md).

## Kiro — research pending

The target user has no admin access. Do not assume that a public Kiro balance API exists. Validate what data the user's own session can read. If a private dashboard interface is used, document the dependency and its fragility, keep regular user permissions and mark it experimental. Never invent quotas from spend or tokens without a known limit.

- [Kiro FAQ](https://kiro.dev/faq/)

The mock examples are for design only. They implement no authentication or scraping of these services.
