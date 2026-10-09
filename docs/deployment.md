# Deployment

**English** · [Español](deployment.es.md)

## Docker Compose / homelab

Follow the README. The `usage-data` volume keeps observations across restarts; `docker compose --env-file server/.env down -v` deletes it. Data is stored per source/account combination: switching from mock to real does not reuse fictitious data.

For the ESP32 to reach the server from your LAN, set `HOST_BIND` to the server's LAN IP or `0.0.0.0` and recreate the service. Use a firewall to limit access to your network. The default value only allows local connections.

Both local Go and Docker use `CONFIG_NAME` and the files in `server/configs/`. The Dockerfile includes the configurations, and Compose mounts that directory so they can be changed without rebuilding the image. `.env` is excluded from the image; secrets are injected through the environment.

## Coolify or cloud

You can deploy `compose.yaml` from Git, set the variables in the panel and keep the volume. Point the platform's proxy at the container's port 8080; this may require removing the `ports` mapping from Compose. No real Coolify deployment has been tested yet.

For connections outside a trusted network, publish HTTPS behind a proxy with a valid certificate. The read/ingestion tokens are still required. Do not open the HTTP port directly to the Internet. The firmware will have to validate the TLS certificate.

The server needs no AI models or GPU; it queries usage data. Codex runs on the server itself. Its session is authorized from a browser with a code and stored in /data/codex. It does not depend on your Mac. The volume contains credentials: back it up and restrict access. The image includes Node/Codex, which is why it is larger than a pure Go image.

Persistence: a single process with an atomic JSON file. Do not scale to several replicas over the same file; a future shared implementation will need another strategy. Back up the volume if you need to keep the data. It stores no history, only the latest observation per account.
