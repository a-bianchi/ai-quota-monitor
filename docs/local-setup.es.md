# Configuración local

[English](local-setup.md) · **Español**

La raíz del repo contiene `compose.yaml`. `server/configs/config.json` usa datos ficticios; `server/configs/config.live.json` incluye OpenRouter directo y Codex personal/laboral directo con login en servidor.

Cada cuenta define `id`, `name`, `provider`, `scope`, `source` y, para OpenRouter directo, `credentialEnv`. IDs únicos en minúsculas, dígitos, guiones y guiones bajos. JSON estricto: campos desconocidos se rechazan.

| Variable | Valor predeterminado / uso |
|---|---|
| `DEVICE_TOKEN` | Obligatoria, mínimo 32 caracteres; lectura de métricas |
| `COLLECTOR_TOKEN` | Opcional; si se define, mínimo 32 caracteres y distinto del anterior |
| `AUTH_USERNAME`, `AUTH_PASSWORD` | Opcionales, van juntas; habilitan `POST /auth/login`. Contraseña de 8+ caracteres; requieren `COLLECTOR_TOKEN` |
| `HTTP_HOST` | `127.0.0.1` sin Docker; Compose usa `0.0.0.0` dentro del contenedor |
| `HTTP_PORT` | `8080`; en Compose configura el puerto del host |
| `HOST_BIND` | Solo Compose: `127.0.0.1` para acceso local |
| `CONFIG_NAME` | Go y Docker: nombre dentro de `server/configs/`, predeterminado `config.json` |
| `MOCK_FILE` | Binario: `configs/mock-usage.json` desde `server/` |
| `DATA_DIR` | Binario: `./data`; Compose: `/data` con volumen persistente |
| `POLL_INTERVAL` | `5m`; duración Go, mínimo 1s |
| `STALE_AFTER` | `15m`; debe ser >= intervalo |
| `OPENROUTER_PERSONAL_KEY` | Management key, solo para configuración live |
| `CURSOR_WORK_SESSION` | Cookie de sesión de cursor.com, solo para configuración live; ver [Cursor](providers/cursor.es.md) |

## OpenRouter real con Compose

Editá `server/.env`: `CONFIG_NAME=config.live.json` y `OPENROUTER_PERSONAL_KEY` con tu clave. Luego `make docker-up`. Las cuentas Codex requieren completar el [login HTTP del servidor](providers/codex.es.md).

Para usar una configuración propia sin versionarla: guardá `server/configs/config.local.json` y elegí `CONFIG_NAME=config.local.json`. Se usa igual con Go y Docker, sin modificar montajes. `CONFIG_NAME` es el nombre del archivo, no una ruta. `CONFIG_FILE` ya no se usa.

El servidor carga `.env` de su directorio de trabajo; `make run` ejecuta desde `server/`. También podés usar `cd server && go run ./cmd/server` con una versión de Go compatible con tu sistema. En esta Mac con Go 1.22 usá `make run`, que resuelve linker/firma.

`make docker-up` equivale a `docker compose --env-file server/.env up --build -d`, desde la raíz. Usamos `--env-file` porque Compose está en la raíz y los secretos pertenecen al servidor.

Variables exportadas en la terminal o definidas por Docker tienen prioridad sobre `.env`; si antes exportaste CONFIG_NAME, usá `unset CONFIG_NAME` para tomar el valor del archivo. El dotenv admite asignaciones simples, comillas y comentarios en líneas completas; no ejecuta comandos ni expande variables. Pegá valores reales, no expresiones de shell.

Los archivos `.env` y `server/data/` están ignorados por Git. `server/.env.example` solo lleva marcadores vacíos. No pegar tokens en issues. El servidor no imprime claves ni cuerpos de errores de proveedores.

## Prueba manual de ingestión

Con `config.live.json` y un token de recolector definido, usá un timestamp UTC actual:

```sh
printf '{"updatedAt":"%s","quotas":[{"id":"5h","name":"5 horas","unit":"percent","used":11,"limit":100,"resetsAt":null}]}' \
  "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > /tmp/quota-report.json
curl -X PUT http://localhost:8080/accounts/codex-work/usage \
  -H 'Authorization: Bearer TU_COLLECTOR_TOKEN' \
  -H 'Content-Type: application/json' \
  --data-binary @/tmp/quota-report.json
```

Esta prueba envía números ficticios, no consulta Codex.

En macOS, `make build` utiliza el linker del sistema y firma localmente el binario (ad-hoc) para compatibilidad con Go 1.22. `make run` compila y arranca ese binario. No se usa una identidad de desarrollador ni se distribuye una firma comercial.

Variables del recolector: `QUOTA_API_URL` (default http://localhost:8080), `CODEX_BIN` (default codex), `CODEX_POLL_INTERVAL` (default 5m, mínimo 10s). No poner credenciales de ChatGPT en `.env`.

`config.live.json` ahora usa source=direct para Codex: la prueba de ingestión anterior requiere una cuenta source=collector en una configuración personalizada. No enviar reportes manuales a las cuentas directas.
