# Cursor

**English** · [Español](cursor.es.md)

Status: **experimental**. It uses the endpoint that feeds the usage panel on cursor.com. Cursor does not document it and it can change without notice; the tests use responses with the shape reported by third parties.

It does not require being an administrator: it only reads the usage of the signed-in account.

## What it queries

`GET https://cursor.com/api/usage-summary`, with the `WorkosCursorSessionToken` cookie. It sends no prompts and consumes no quota.

## Configuration

1. Open <https://cursor.com/dashboard> with the right account.
2. DevTools → Application → Cookies → `https://cursor.com` → copy the value of `WorkosCursorSessionToken`.
3. Paste it in `server/.env`:

```env
CURSOR_WORK_SESSION=cookie_value
```

4. Declare the account in `server/configs/`:

```json
{
  "id": "cursor-work",
  "name": "Cursor trabajo",
  "provider": "cursor",
  "scope": "work",
  "source": "direct",
  "credentialEnv": "CURSOR_WORK_SESSION"
}
```

5. Restart the server and query `GET /usage`.

The cookie is a credential: it gives access to your Cursor account. Do not commit or share it. For another account, add another entry with another variable; with Docker Compose you also have to pass it in `compose.yaml`.

## Returned quotas

All in `percent`, with `resetsAt` equal to the end of the billing cycle:

| ID | Meaning | Source |
|---|---|---|
| `included` | Usage included in the plan | `totalPercentUsed`, or `used/limit` from `individualUsage.plan` or `overall` |
| `auto` | Share from Auto models | `autoPercentUsed`, if present |
| `api` | Share from manually chosen models | `apiPercentUsed`, if present |
| `on-demand` | Your own on-demand spend against its cap | `individualUsage.onDemand`, if enabled and it has a limit |
| `team-on-demand` | Team on-demand spend against its cap | `teamUsage.onDemand`, if enabled and it has a limit |

Only meters with a known limit are reported. Percentages are used because the unit of `used`/`limit` is not documented (requests or cents depending on the plan); they are not converted to USD.

## Known limitations

- The session expires and does not renew itself. When that happens, the account goes to `auth_required`: copy the cookie again and restart.
- Some Enterprise accounts expose no meter with a limit; in that case the account stays `unavailable`.
- Your organization's policies may limit what the dashboard shows.
