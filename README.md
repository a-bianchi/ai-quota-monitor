# AI Quota Monitor

[English](README.en.md) · **Español**

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

## Integraciones

| Servicio | Estado del repo | Fuente |
|---|---|---|
| OpenRouter | Cliente implementado | API oficial de créditos; requiere management key |
| Codex / Work | Integrado en servidor; login HTTP por código y polling automático | App Server `account/rateLimits/read` |
| Claude Pro/Max | Integrado en servidor, experimental; [no permitido por los términos de Anthropic](docs/providers/claude-subscription.md) | Login OAuth y endpoint de uso no documentado |
| Cursor | Cliente implementado, experimental; sin validar con cuenta real | Endpoint no documentado del dashboard (`/api/usage-summary`) con cookie de sesión |
| Kiro | Investigación pendiente con acceso sin admin | Panel de créditos; vía automática por validar |
| ChatGPT normal | No integrado | No se confirmó un contador completo por API |

Para conectar tus cuentas, seguí la [guía paso a paso](docs/providers/conectar.md). Ver también [proveedores](docs/providers/README.md). Los cupos de Codex no representan todas las funciones de ChatGPT.

## Consultar Codex desde el homelab

Con `CONFIG_NAME=config.live.json`, la API consulta Codex directamente. No necesita un proceso en tu Mac. Usá tu `COLLECTOR_TOKEN` existente (32+ caracteres y distinto de DEVICE_TOKEN) en `server/.env`, reiniciá la API y enviá:

```sh
curl -X POST http://localhost:8080/accounts/codex-work/codex/login \
  -H 'Authorization: Bearer TU_COLLECTOR_TOKEN'
```

La respuesta devuelve `verificationUrl` y `userCode`. Abrí esa URL en tu navegador, ingresá el código y elegí la cuenta laboral. La sesión queda en el servidor. Consultá el estado con GET sobre la misma URL y el mismo token. Al completarse el login se actualizan los cupos automáticamente.

Para tu cuenta personal usá `codex-personal`. Ver [guía de login del servidor](docs/providers/codex.md).

## Arquitectura

```text
OpenRouter + Codex ──> servidor Go ──> GET /usage ──> ESP32
                            ↑
Navegador ── autoriza login por código de Codex
```

El servidor consulta fuentes directas de OpenRouter/Codex y conserva la ingestión opcional para otras cuentas configuradas como `collector`. Guarda el último valor; señala datos viejos o problemas de autenticación. Las lecturas del ESP32 no disparan nuevas consultas al proveedor.

## Documentación

- [Contrato y endpoints](docs/api.md)
- [Configuración local y variables](docs/local-setup.md)
- [Docker, homelab, Coolify y nube](docs/deployment.md)
- [Hardware y lista de materiales](docs/hardware.md)
- [Firmware y diseño de pantalla](docs/firmware.md)
- [Carcasa e impresión 3D](enclosure/README.md)
- [Roadmap](docs/roadmap.md)

Configuración JSON en `server/configs/`; secretos en `server/.env` o variables de entorno. Esta primera versión usa exclusivamente la biblioteca estándar de Go. El modo mock sirve para desarrollar sin hardware.

## Contribuir

Ejecutá `make test` y `cd server && go vet ./...` antes de proponer cambios. Cada integración debe documentar autenticación, unidades, actualización y limitaciones. No incluir cookies, archivos de sesión, claves o información del trabajo en commits ni issues.

Licencia MIT. No afiliado a los fabricantes ni proveedores mencionados.
