# Deployment

**English** · [Español](deployment.es.md)

## Docker Compose / homelab

Requirements: Git and Docker with Compose v2. From the repo root:

```sh
cp server/.env.example server/.env
# Fill in DEVICE_TOKEN with the output of: openssl rand -hex 32
make docker-up
curl http://localhost:8080/ping
```

`make docker-up` is equivalent to `docker compose --env-file server/.env up --build -d`. `--env-file` is used because `compose.yaml` lives at the root and the secrets belong to the server. To stop it, `make docker-down`; to see the logs, `docker compose --env-file server/.env logs -f api`.

The variables are described in [configuration](configuration.md). The `usage-data` volume keeps observations across restarts; `docker compose --env-file server/.env down -v` deletes it. Data is stored per source/account combination: switching from mock to real does not reuse fictitious data.

For the ESP32 to reach the server from your LAN, set `HOST_BIND` to the server's LAN IP or `0.0.0.0` and recreate the service. Use a firewall to limit access to your network. The default value only allows local connections.

```dotenv
# server/.env — example; use your server's LAN IP
HOST_BIND=192.168.1.50
```

After changing it, run `make docker-up` again and test from another machine with `curl http://192.168.1.50:8080/ping`.

Both local Go and Docker use `CONFIG_NAME` and the files in `server/configs/`. The Dockerfile includes the configurations, and Compose mounts that directory so they can be changed without rebuilding the image. `.env` is excluded from the image; secrets are injected through the environment.

## Web panel

The panel is part of the same binary and container: it is served at `/` on the same port as the API, with nothing else to deploy. Set `AUTH_USERNAME`, `AUTH_PASSWORD_HASH` and `MASTER_KEY` in `server/.env` (or in the platform's variables); Compose passes them to the container. See [web panel](panel.md) to create the user.

Panel sessions (`panel-sessions.json`) and encrypted provider credentials (`credentials.json`) live in the `usage-data` volume next to the Codex and Claude sessions. Keep `MASTER_KEY` outside that volume: a copy of the volume alone must not be enough to read the credentials. Losing `MASTER_KEY` means loading them again.

Reach the panel over HTTPS whenever it is not on a trusted network. `HOST_BIND` controls who can connect, as for the API.

## Coolify or cloud

You can deploy `compose.yaml` from Git, set the variables in the panel and keep the volume. Point the platform's proxy at the container's port 8080; this may require removing the `ports` mapping from Compose. No real Coolify deployment has been tested yet.

For connections outside a trusted network, publish HTTPS behind a proxy with a valid certificate. The read/ingestion tokens are still required. Do not open the HTTP port directly to the Internet. The firmware will have to validate the TLS certificate.

The server needs no AI models or GPU; it queries usage data. Codex runs on the server itself. Its session is authorized from a browser with a code and stored in /data/codex. It does not depend on your Mac. The volume contains credentials: back it up and restrict access. The image includes Node/Codex, which is why it is larger than a pure Go image.

Persistence: a single process with an atomic JSON file. Do not scale to several replicas over the same file; a future shared implementation will need another strategy. Back up the volume if you need to keep the data. It stores no history, only the latest observation per account.
