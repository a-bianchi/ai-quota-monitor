# Contributing

**English** · [Español](CONTRIBUTING.es.md)

Thanks for your interest in the project. This guide explains how to run it, test it and propose changes.

## Running the project

Requirements: Git, Go 1.22 or later, `make` and `openssl`. From the repo root:

```sh
cp server/.env.example server/.env
# Paste into DEVICE_TOKEN the output of: openssl rand -hex 32
make run
```

The server starts at `http://localhost:8080` with example data, without any real account. For the web panel, create a user with `make panel-password`. More detail in [local setup](docs/local-setup.md).

No provider credential is needed for development: example mode and the tests use simulated data and servers.

## Running the tests

```sh
make test
cd server && go vet ./...
```

`make test` runs `go test -race ./...` inside `server/`. The GitHub Actions workflow (`.github/workflows/go.yml`) runs those same two steps on every push and pull request.

Before submitting a change, format the code with `gofmt`:

```sh
cd server && gofmt -l .
```

The command must not list any file.

## Layout

The code lives in `server/`; [architecture](docs/architecture.md) describes each package. The server uses only the Go standard library: do not add external dependencies without discussing it in an issue first.

## Adding a provider

Each integration lives in its own package, `server/internal/providers/<provider>/`. Use `openrouter/` or `cursor/` as a reference: they are the simplest.

1. **Create the package** with a client that returns a `usage.Report`:

   ```go
   func (p Provider) Collect(ctx context.Context, credential string) (usage.Report, error)
   ```

   - Convert the provider's response to quotas (`usage.Quota`) with `unit` set to `percent`, `credits`, `usd` or `tokens`.
   - Validate the result with `usage.ValidateReport` before returning it.
   - Export an `ErrAuth` error and return it when the credential is missing, rejected or expired.
   - Use timeouts, do not log credentials or response bodies, and do not send prompts or consume quota.
   - Report only meters with a known limit: do not invent quotas from spend without a limit.

2. **Register the provider** in `server/internal/config/config.go`: add it to the list of valid `provider` values and, if it supports `source: direct`, to that `switch`, stating whether it requires `credentialEnv`.

3. **Wire it into the polling loop** in `server/cmd/server/main.go`: add its case to the `switch a.Provider` in `collect` and add its `ErrAuth` to the condition that produces `auth_required`.

4. **Add the panel texts** in `server/internal/web/static/i18n/es.json` and `en.json`: the name under `provider.<provider>` and, if it uses a credential, its label under `settings.credential.field.<provider>`. Quota names are written in Spanish; their English translation goes in `server/internal/httpapi/localize.go`.

5. **Write tests** with `httptest`, covering the successful response, the authentication error and incomplete responses. Do not use real credentials or call the real service in tests.

6. **Document it** in `docs/providers/<provider>.md` and `docs/providers/<provider>.es.md`: authentication, endpoint queried, units, quotas returned, reset behavior and known limits. State clearly whether the endpoint is official and whether it was validated with a real account. Update the tables in `docs/providers/README.md`, `docs/providers/connect.md` and the main README, in both languages.

7. **If it uses an environment variable**, add it empty to `server/.env.example` and to the `environment` section of `compose.yaml`.

If the provider needs an interactive login, like Codex or Claude, look at those packages and their routes in `server/internal/httpapi/api.go`. Open an issue first to agree on the design.

## Conventions

- **Small, focused changes**: one pull request per topic.
- **Commit messages** are short, with a prefix for the type: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`.
- **Code** follows the style of the file you touch, formatted with `gofmt`. Identifiers, comments and error messages are in English.
- **Tests** for every change in behavior.
- **Documentation** in English (`file.md`) and Spanish (`file.es.md`), with links between both versions at the top. If you change one, update the other.
- **No invented features**: document what works and mark as pending or experimental what has not been validated.

## Pull requests

1. Fork the repo and create a branch from `main`.
2. Make the change with its tests and documentation.
3. Run `make test` and `cd server && go vet ./...`.
4. Open the pull request explaining what changes, why, and how you tested it.

## Security and private data

- Do not include tokens, keys, cookies, session files, emails or work information in commits, issues, screenshots or logs.
- `server/.env`, `server/data/` and `server/configs/config.local.json` are in `.gitignore`: do not force them in.
- Screenshots must use example mode.
- If you find a security problem, do not publish the details in an issue: contact the repository maintainer first.

## License

By contributing you agree that your contribution is published under the project's [MIT license](LICENSE).
