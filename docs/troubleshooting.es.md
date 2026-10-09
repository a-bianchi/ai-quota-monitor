# Solución de problemas

[English](troubleshooting.md) · **Español**

Primero mirá los logs: el servidor explica por qué no arranca o qué cuenta falló.

- Local: la salida de `make run`.
- Docker, desde la raíz del repo: `docker compose --env-file server/.env logs api`.

Los logs no incluyen keys ni cuerpos de respuestas de proveedores.

## El servidor no arranca

El log muestra `startup failed` con el motivo.

| Mensaje | Causa y solución |
|---|---|
| `DEVICE_TOKEN must contain at least 32 characters` | Falta `DEVICE_TOKEN` en `server/.env` o es corto. Generalo con `openssl rand -hex 32` |
| `COLLECTOR_TOKEN must contain at least 32 characters and differ from DEVICE_TOKEN` | Usá un segundo token, distinto del primero |
| `set both AUTH_USERNAME and AUTH_PASSWORD_HASH, or neither` | Definiste solo una de las dos variables del panel |
| `AUTH_PASSWORD_HASH is not valid` | El hash está incompleto o modificado. Volvé a generarlo con `make panel-password` y pegá la línea entera, sin comillas |
| `AUTH_PASSWORD is no longer supported` | Borrá `AUTH_PASSWORD` y usá `AUTH_PASSWORD_HASH` |
| `MASTER_KEY must contain at least 32 characters` | Generala con `openssl rand -base64 32` |
| `STALE_AFTER must be a duration >= POLL_INTERVAL` | Subí `STALE_AFTER` o bajá `POLL_INTERVAL` |
| `invalid dotenv assignment at line N` | Esa línea de `server/.env` no tiene la forma `NOMBRE=valor` |
| `direct openrouter requires credentialEnv`, `unknown provider`, `invalid or duplicate account ID/name` | Error en el archivo de cuentas. Ver [configuración](configuration.es.md#cuentas) |
| `invalid or missing mock data for <id>` | Una cuenta `mock` no tiene entrada en `server/configs/mock-usage.json` |
| `address already in use` | Otro proceso usa el puerto. Cambiá `HTTP_PORT` o detené el otro proceso; `make run` y `make docker-up` no pueden compartir el 8080 |

Con Compose, `Set DEVICE_TOKEN in .env` significa que falta el token o que no se pasó `--env-file server/.env`. Usá `make docker-up`.

## Un cambio en `.env` no tiene efecto

- Reiniciá el servidor: las variables se leen solo al arrancar.
- Una variable exportada en la terminal le gana al archivo. Probá `unset NOMBRE_DE_LA_VARIABLE`.
- `make run` ejecuta desde `server/`, por eso el archivo debe ser `server/.env`.

## Respuestas de error de la API

| Código | Significado |
|---|---|
| 401 `unauthorized` | Token incorrecto o falta el encabezado `Authorization: Bearer <token>`. Revisá espacios o comillas sobrantes |
| 503 `endpoint disabled` | La ruta depende de algo sin configurar: `COLLECTOR_TOKEN` para la ingestión y los logins por token, o el usuario del panel |
| 400 `invalid X-Timezone` | Usá un nombre IANA, por ejemplo `America/Argentina/Buenos_Aires` o `UTC` |
| 404 en `PUT /accounts/{id}/usage` | La cuenta no existe o no es `source: collector` |
| 409 en la ingestión | El reporte es más viejo que el dato guardado |

## El panel

- **"El login del panel no está configurado en el servidor"**: faltan `AUTH_USERNAME` y `AUTH_PASSWORD_HASH`, o no reiniciaste. Ver [crear el usuario](panel.es.md#crear-el-usuario).
- **"Usuario o contraseña incorrectos"**: se ingresa la contraseña que tipeaste en `make panel-password`, no el hash. Si la olvidaste, generá un hash nuevo.
- **"Demasiados intentos fallidos"**: después de 5 fallos el login se bloquea 15 minutos.
- **"El servidor no tiene MASTER_KEY"**: es un aviso, no un error. Sin `MASTER_KEY` no se pueden guardar credenciales desde el panel; las variables de entorno siguen funcionando.
- **Las credenciales guardadas desaparecieron**: cambió `MASTER_KEY` o se borró la carpeta de datos. Hay que cargarlas de nuevo.

## Una cuenta no muestra datos

Consultá `GET /usage` y mirá el `status` de la cuenta.

| `status` | Qué hacer |
|---|---|
| `auth_required` | Falta la credencial o venció. Repetí los pasos de conexión de ese proveedor en la [guía](providers/connect.es.md) |
| `unavailable` | Nunca hubo una lectura correcta. Revisá los logs, la conexión a Internet del servidor y, para Codex, que `CODEX_BIN` apunte a un ejecutable válido |
| `stale` | Hay datos, pero viejos. Esperá un ciclo de `POLL_INTERVAL`; si sigue, revisá los logs |

Casos por proveedor:

- **Codex**: el login responde 502 si Codex CLI no está instalado o si la organización restringe el login por código de dispositivo. Ver [Codex](providers/codex.es.md).
- **Cursor**: la cookie vence y no se renueva sola. Copiala de nuevo. Ver [Cursor](providers/cursor.es.md).
- **Claude**: si el código es rechazado o pasaron 15 minutos, empezá el login otra vez. Ver [Claude](providers/claude-subscription.es.md).
- **OpenRouter**: hace falta una management key; una key de inferencia no sirve.

## Docker y red local

- **No se puede entrar desde otro equipo de la red**: por defecto el puerto se publica solo en `127.0.0.1`. Poné `HOST_BIND` con la IP LAN del servidor y ejecutá `make docker-up` otra vez. Ver [despliegue](deployment.es.md).
- **Una variable nueva no llega al contenedor**: `compose.yaml` pasa una lista fija de variables. Agregala en su sección `environment`.
- **Los datos de ejemplo siguen apareciendo**: revisá `CONFIG_NAME`. Los datos se guardan por `source` e `id`, así que una cuenta `mock` y una `direct` no comparten valores.
- **Empezar de cero**: `docker compose --env-file server/.env down -v` borra el volumen, incluidas las sesiones de Codex y Claude.

## macOS

Si `go build` o `go test` fallan con errores del linker, usá `make build` y `make test`, que pasan los parámetros necesarios y firman el binario localmente.
