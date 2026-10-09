FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# Codex native package uses glibc; keep the runtime on Debian, not Alpine.
FROM node:22-bookworm-slim
ARG CODEX_VERSION=0.162.0
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && npm install --global @openai/codex@${CODEX_VERSION} \
    && npm cache clean --force \
    && groupadd -g 10001 app && useradd -m -u 10001 -g app app \
    && mkdir /data && chown app:app /data
COPY --from=build /out/server /usr/local/bin/server
WORKDIR /app
COPY configs /app/configs
USER app
EXPOSE 8080
# The image has no curl; Node is already there for Codex.
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["node", "-e", "fetch('http://127.0.0.1:8080/ping').then(r=>process.exit(r.ok?0:1),()=>process.exit(1))"]
ENTRYPOINT ["server"]
