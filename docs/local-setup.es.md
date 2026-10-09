# Instalación y ejecución local

[English](local-setup.md) · **Español**

Cómo ejecutar el servidor en tu máquina sin Docker. Para Docker Compose, ver [despliegue](deployment.es.md).

## Requisitos

- Git.
- Go 1.22 o posterior.
- `make` y `openssl`.
- En macOS, Xcode Command Line Tools: `make build` usa el linker del sistema y firma el binario localmente (firma ad-hoc, sin identidad de desarrollador).
- Codex CLI, solo si vas a usar cuentas de Codex. Se indica con `CODEX_BIN`.

El servidor usa únicamente la biblioteca estándar de Go: no hay dependencias que descargar.

## Primer arranque

Todos los comandos se ejecutan desde la raíz del repo.

1. Creá el archivo de entorno:

   ```sh
   cp server/.env.example server/.env
   ```

2. Generá un token y pegalo en `DEVICE_TOKEN` dentro de `server/.env`:

   ```sh
   openssl rand -hex 32
   ```

3. Compilá y arrancá:

   ```sh
   make run
   ```

4. En otra terminal, comprobá que responde:

   ```sh
   curl http://localhost:8080/ping
   curl -H 'Authorization: Bearer TU_DEVICE_TOKEN' http://localhost:8080/usage
   ```

El servidor escucha en `127.0.0.1:8080` y arranca en modo de ejemplo: seis cuentas ficticias con `source: mock`. `used` es el consumo; el saldo es `limit - used`.

Los datos se guardan en `server/data/`, que está en `.gitignore` junto con `server/.env`.

## Comandos de `make`

| Comando | Qué hace |
|---|---|
| `make run` | Compila y arranca el servidor desde `server/` |
| `make build` | Compila `server/bin/ai-quota-monitor` |
| `make test` | Corre los tests con el detector de carreras |
| `make panel-password` | Pide una contraseña e imprime la línea `AUTH_PASSWORD_HASH=...` |
| `make docker-up` / `make docker-down` | Levanta o detiene el servicio con Docker Compose |
| `make build-collector` | Compila el recolector opcional de Codex |

También podés usar `cd server && go run ./cmd/server`. En macOS conviene `make run`, que resuelve el linker y la firma.

## Siguientes pasos

- **Panel web**: creá el usuario con `make panel-password` y abrí <http://localhost:8080/>. Ver [panel web](panel.es.md).
- **Cuentas reales**: poné `CONFIG_NAME=config.live.json` en `server/.env` y seguí la [guía de proveedores](providers/connect.es.md).
- **Variables y archivo de cuentas**: ver [configuración](configuration.es.md).

## Probar la ingestión a mano

La ingestión solo acepta cuentas con `source: collector`, y las configuraciones incluidas no traen ninguna. Para probarla:

1. Guardá `server/configs/config.local.json` con una cuenta de ese tipo:

   ```json
   {"accounts": [{"id": "codex-work", "name": "Codex trabajo", "provider": "openai", "scope": "work", "source": "collector"}]}
   ```

2. En `server/.env`, poné `CONFIG_NAME=config.local.json` y un `COLLECTOR_TOKEN` generado con `openssl rand -hex 32`. Reiniciá con `make run`.

3. Enviá un reporte con la hora UTC actual:

   ```sh
   printf '{"updatedAt":"%s","quotas":[{"id":"5h","name":"5 horas","unit":"percent","used":11,"limit":100,"resetsAt":null}]}' \
     "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > /tmp/quota-report.json
   curl -X PUT http://localhost:8080/accounts/codex-work/usage \
     -H 'Authorization: Bearer TU_COLLECTOR_TOKEN' \
     -H 'Content-Type: application/json' \
     --data-binary @/tmp/quota-report.json
   ```

Son números ficticios: esta prueba no consulta Codex. No envíes reportes a cuentas `direct`; el servidor los rechaza con 404.

## Recolector opcional de Codex

`server/cmd/collector` lee los cupos de Codex en tu máquina y los envía a la API. Solo hace falta si declarás una cuenta `openai` con `source: collector`; con `config.live.json` el servidor consulta Codex directamente y no se usa.

```sh
make codex-login ACCOUNT=codex-work     # inicia sesión en un perfil aislado
make collect-codex-once ACCOUNT=codex-work
make collect-codex ACCOUNT=codex-work   # envía cada CODEX_POLL_INTERVAL
```

Sus variables están en [configuración](configuration.es.md#recolector-opcional). No pongas credenciales de ChatGPT en `.env`.
