# AI Quota Monitor

[English](README.md) · **Español**

Monitor de escritorio para ver el saldo de servicios de IA, con páginas separadas para **Personal** y **Trabajo**, usando una Waveshare ESP32-C6-LCD-1.47.

**Estado: MVP del servidor.** API Go, modo de ejemplo, persistencia y cliente OpenRouter implementados. Codex integrado en el servidor con login por código de dispositivo y perfiles aislados. Firmware, otros recolectores y carcasa todavía pendientes. No se publican datos ni credenciales de cuentas reales.

## Empezar con Docker Compose

Requisitos: Git y Docker con Compose v2.

```sh
cp server/.env.example server/.env
openssl rand -hex 32
```

Copiá el resultado en `DEVICE_TOKEN` dentro de `server/.env`. Generá otro token distinto para `COLLECTOR_TOKEN` si vas a recibir métricas de un recolector. Luego:

```sh
make docker-up
curl http://localhost:8080/healthz
```

Para consultar datos, reemplazá el marcador por tu token:

```sh
curl -H 'Authorization: Bearer TU_DEVICE_TOKEN' \
  'http://localhost:8080/usage?scope=personal'
```

La configuración inicial muestra **seis cuentas ficticias**, con `source: mock`. `used` significa consumo; la pantalla debe calcular el saldo restante. Los ejemplos no prueban acceso real a Cursor, Kiro ni Claude. Puede tomar unos instantes cargar el primer lote.

## Ejecutar sin Docker

Requisito: Go 1.22 o posterior; en macOS, Xcode Command Line Tools para el linker del sistema. Desde la raíz del repo:

```sh
cp server/.env.example server/.env
# Completá DEVICE_TOKEN en server/.env con un token generado con openssl rand -hex 32.
make run
```

El servidor escucha en `127.0.0.1:8080`. Los datos se guardan en `server/data/`. El servidor carga automáticamente `server/.env` cuando se ejecuta desde `server/` (como hace `make run`). Variables exportadas en la terminal tienen prioridad. Docker usa el mismo archivo mediante `make docker-up`.

```sh
make test
make build
```

Para usar OpenRouter real, poné `CONFIG_NAME=config.live.json` y tu management key en `server/.env`. Reiniciá con `make run` o `make docker-up`.

## Panel web

La API también sirve un panel web en <http://localhost:8080/>: tablero de servicios, vista previa de la pantalla del ESP32 y configuración para iniciar sesión en Codex y cargar keys de proveedores. Va incluido en el servidor (sin Node ni paso de build) y queda deshabilitado hasta que crees un usuario:

```sh
make panel-password        # pide una contraseña e imprime AUTH_PASSWORD_HASH=...
openssl rand -base64 32    # valor para MASTER_KEY
```

Poné `AUTH_USERNAME`, la línea `AUTH_PASSWORD_HASH` y `MASTER_KEY` en `server/.env`, reiniciá y abrí la URL. Detalle en [panel web](docs/panel.es.md).

## Integraciones

| Servicio | Estado del repo | Fuente |
|---|---|---|
| OpenRouter | Cliente implementado | API oficial de créditos; requiere management key |
| Codex / Work | Integrado en servidor; login HTTP por código y polling automático | App Server `account/rateLimits/read` |
| Claude Pro/Max | Integrado en servidor, experimental; [no permitido por los términos de Anthropic](docs/providers/claude-subscription.es.md) | Login OAuth y endpoint de uso no documentado |
| Cursor | Cliente implementado, experimental; sin validar con cuenta real | Endpoint no documentado del dashboard (`/api/usage-summary`) con cookie de sesión |
| Kiro | Investigación pendiente con acceso sin admin | Panel de créditos; vía automática por validar |
| ChatGPT normal | No integrado | No se confirmó un contador completo por API |

Para conectar tus cuentas, seguí la [guía paso a paso](docs/providers/connect.es.md). Ver también [proveedores](docs/providers/README.es.md). Los cupos de Codex no representan todas las funciones de ChatGPT.

## Consultar Codex desde el homelab

Con `CONFIG_NAME=config.live.json`, la API consulta Codex directamente. No necesita un proceso en tu Mac. Usá tu `COLLECTOR_TOKEN` existente (32+ caracteres y distinto de DEVICE_TOKEN) en `server/.env`, reiniciá la API y enviá:

```sh
curl -X POST http://localhost:8080/accounts/codex-work/codex/login \
  -H 'Authorization: Bearer TU_COLLECTOR_TOKEN'
```

La respuesta devuelve `verificationUrl` y `userCode`. Abrí esa URL en tu navegador, ingresá el código y elegí la cuenta laboral. La sesión queda en el servidor. Consultá el estado con GET sobre la misma URL y el mismo token. Al completarse el login se actualizan los cupos automáticamente.

Para tu cuenta personal usá `codex-personal`. Ver [guía de login del servidor](docs/providers/codex.es.md).

## Arquitectura

```text
OpenRouter + Codex ──> servidor Go ──> GET /usage ──> ESP32
                            ↑
Navegador ── autoriza login por código de Codex
```

El servidor consulta fuentes directas de OpenRouter/Codex y conserva la ingestión opcional para otras cuentas configuradas como `collector`. Guarda el último valor; señala datos viejos o problemas de autenticación. Las lecturas del ESP32 no disparan nuevas consultas al proveedor.

## Colección de Postman

`docs/collections/` contiene una colección y un environment de Postman listos para usar:

- `ai-quota-monitor.postman_collection.json`: login, refresh, `GET /usage` y los logins de Codex y Claude.
- `ai-quota-monitor.postman_environment.json`: las variables que usan las peticiones, con valores vacíos.

Importá los dos archivos en Postman (**Import** → arrastrá ambos), elegí el environment `ai-quota-monitor` y completá `url` (por ejemplo `http://localhost:8080`), `COLLECTOR_TOKEN` y, si habilitaste el login, `username` y `password`. Las peticiones de login y refresh guardan `access_token` y `refresh_token` en el environment automáticamente. Dejá tus valores en Postman; no los subas al repo.

## Documentación

- [Panel web](docs/panel.es.md)
- [Contrato y endpoints](docs/api.es.md)
- [Colección y environment de Postman](docs/collections/)
- [Configuración local y variables](docs/local-setup.es.md)
- [Docker, homelab, Coolify y nube](docs/deployment.es.md)
- [Hardware y lista de materiales](docs/hardware.es.md)
- [Firmware y diseño de pantalla](docs/firmware.es.md)
- [Carcasa e impresión 3D](enclosure/README.es.md)
- [Roadmap](docs/roadmap.es.md)

Configuración JSON en `server/configs/`; secretos en `server/.env` o variables de entorno. El servidor usa solo la biblioteca estándar de Go; el panel incluye Alpine.js localmente. El modo mock sirve para desarrollar sin hardware.

## Contribuir

Ejecutá `make test` y `cd server && go vet ./...` antes de proponer cambios. Cada integración debe documentar autenticación, unidades, actualización y limitaciones. No incluir cookies, archivos de sesión, claves o información del trabajo en commits ni issues.

Licencia MIT. No afiliado a los fabricantes ni proveedores mencionados.
