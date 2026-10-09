# Configuración

[English](configuration.md) · **Español**

El servidor se configura con dos cosas:

- **Variables de entorno**, en `server/.env`: tokens, secretos y ajustes del proceso.
- **Un archivo de cuentas**, en `server/configs/`: qué servicios se monitorean y si son personales o laborales.

Después de cambiar cualquiera de los dos hay que reiniciar el servidor.

## Variables de entorno

Copiá el archivo de ejemplo desde la raíz del repo y completalo:

```sh
cp server/.env.example server/.env
```

El servidor lee `.env` de su directorio de trabajo; `make run` lo ejecuta desde `server/`. `make docker-up` pasa el mismo archivo a Compose. El archivo está en `.gitignore`.

### Tokens y acceso

| Variable | Obligatoria | Propósito | Ejemplo |
|---|---|---|---|
| `DEVICE_TOKEN` | Sí | Token de lectura para `GET /usage`; es el que usa el display. Mínimo 32 caracteres | Salida de `openssl rand -hex 32` |
| `COLLECTOR_TOKEN` | No | Habilita la ingestión y permite iniciar los logins de Codex y Claude por HTTP. Mínimo 32 caracteres y distinto de `DEVICE_TOKEN` | Salida de `openssl rand -hex 32` |
| `AUTH_USERNAME` | No | Usuario del [panel web](panel.es.md). Va junto con `AUTH_PASSWORD_HASH` | `admin` |
| `AUTH_PASSWORD_HASH` | No | Hash de la contraseña del panel. Se genera con `make panel-password` | `pbkdf2-sha256:600000:...` |
| `MASTER_KEY` | No | Cifra las credenciales de proveedores guardadas desde el panel. Mínimo 32 caracteres | Salida de `openssl rand -base64 32` |

`AUTH_PASSWORD` en texto plano no se acepta: si está definida, el servidor no arranca.

### Servidor

| Variable | Valor por defecto | Propósito |
|---|---|---|
| `CONFIG_NAME` | `config.json` | Nombre del archivo de cuentas dentro de `server/configs/`. Es un nombre, no una ruta |
| `HTTP_PORT` | `8080` | Puerto. En Compose es el puerto publicado en el host |
| `HTTP_HOST` | `127.0.0.1` | Dirección de escucha del binario. Compose la fija en `0.0.0.0` dentro del contenedor |
| `HOST_BIND` | `127.0.0.1` | Solo Compose: interfaz del host donde se publica el puerto |
| `POLL_INTERVAL` | `5m` | Cada cuánto se consulta a los proveedores. Duración Go, mínimo `1s` |
| `STALE_AFTER` | `15m` | Antigüedad a partir de la cual una cuenta pasa a `stale`. Debe ser mayor o igual a `POLL_INTERVAL` |
| `DATA_DIR` | `./data` | Carpeta de datos. Compose la fija en `/data`, sobre el volumen `usage-data` |
| `MOCK_FILE` | `configs/mock-usage.json` | Datos de las cuentas `mock` |
| `CODEX_BIN` | `codex` | Ejecutable de Codex CLI. La imagen Docker ya lo incluye |

### Credenciales de proveedores

El nombre de estas variables lo define el campo `credentialEnv` de cada cuenta. Las de `config.live.json` son:

| Variable | Propósito |
|---|---|
| `OPENROUTER_PERSONAL_KEY` | Management key de OpenRouter |
| `CURSOR_WORK_SESSION` | Valor de la cookie `WorkosCursorSessionToken` de cursor.com |

Son opcionales si cargás la credencial desde el panel; lo guardado en el panel tiene prioridad. Codex y Claude no usan variables: inician sesión por HTTP o desde el panel.

Con Docker Compose, `compose.yaml` pasa al contenedor una lista fija de variables. Si agregás una cuenta con un `credentialEnv` nuevo, sumá esa variable en la sección `environment` de `compose.yaml` o cargá la credencial desde el panel.

### Recolector opcional

Solo para `server/cmd/collector`, que envía los cupos de una cuenta Codex `source: collector`:

| Variable | Valor por defecto | Propósito |
|---|---|---|
| `QUOTA_API_URL` | `http://localhost:8080` | URL de la API que recibe los datos |
| `CODEX_POLL_INTERVAL` | `5m` | Intervalo de envío, mínimo `10s` |

### Reglas del archivo `.env`

- Las variables exportadas en la terminal o definidas por Docker tienen prioridad sobre `.env`. Si exportaste `CONFIG_NAME` antes, usá `unset CONFIG_NAME`.
- Admite asignaciones simples, comillas y comentarios en líneas completas.
- No ejecuta comandos ni expande variables: pegá valores reales, no expresiones de shell.

## Tokens de la API

Hay tres formas de autenticarse y son independientes entre sí:

| Credencial | Quién la usa | Qué permite |
|---|---|---|
| `DEVICE_TOKEN` | El display y cualquier cliente de solo lectura | `GET /usage` |
| `COLLECTOR_TOKEN` | Recolectores y administración por línea de comandos | `GET /usage`, ingestión y logins de Codex y Claude |
| Sesión del panel | El navegador, después de iniciar sesión | `GET /usage`, logins de Codex y Claude y carga de credenciales |

Todas se envían como `Authorization: Bearer <token>`. La tabla completa por ruta está en [panel web](panel.es.md#sesión) y los endpoints en la [referencia de la API](api.es.md).

## Cuentas

El archivo de cuentas es un JSON con una lista `accounts`. El repo trae dos:

- `server/configs/config.json`: seis cuentas ficticias (`source: mock`). Es el valor por defecto.
- `server/configs/config.live.json`: cuentas reales de OpenRouter, Cursor, Codex personal y laboral, y Claude.

Para una configuración propia sin versionarla, guardala como `server/configs/config.local.json` (está en `.gitignore`) y poné `CONFIG_NAME=config.local.json`. Funciona igual con Go y con Docker.

```json
{
  "accounts": [
    {"id": "openrouter-personal", "name": "OpenRouter", "provider": "openrouter", "scope": "personal", "source": "direct", "credentialEnv": "OPENROUTER_PERSONAL_KEY"},
    {"id": "codex-work", "name": "Codex trabajo", "provider": "openai", "scope": "work", "source": "direct"}
  ]
}
```

| Campo | Valores | Notas |
|---|---|---|
| `id` | Minúsculas, dígitos, `-` y `_`; hasta 80 caracteres | Único. Aparece en las URLs y en las carpetas de sesión |
| `name` | Texto, hasta 120 caracteres | Nombre que se muestra |
| `provider` | `openai`, `openrouter`, `cursor`, `anthropic`, `kiro` | `openai` es Codex; `anthropic` es Claude |
| `scope` | `personal`, `work` | Separa las dos páginas del panel y del display |
| `source` | `mock`, `direct`, `collector` | De dónde salen los datos |
| `credentialEnv` | Nombre de una variable de entorno | Obligatorio para `openrouter` y `cursor` con `source: direct` |

Reglas:

- Entre 1 y 64 cuentas. El JSON es estricto: un campo desconocido impide arrancar.
- `source: direct` existe para `openai`, `openrouter`, `cursor` y `anthropic`. `kiro` solo admite `mock` o `collector`.
- Cada cuenta `mock` necesita una entrada con su `id` en `server/configs/mock-usage.json`.
- Una cuenta `collector` no se consulta: espera datos por [ingestión](api.es.md#ingestión).
- Los datos se guardan por combinación de `source` e `id`: pasar una cuenta de `mock` a `direct` no reutiliza los valores ficticios.

### Personal y laboral

Para tener el mismo servicio en los dos ámbitos, declaralo dos veces con `id` y `scope` distintos, como `codex-personal` y `codex-work` en `config.live.json`. Cada cuenta guarda su propia sesión o credencial. El `id` no comprueba qué cuenta elegiste al iniciar sesión en el navegador: usá perfiles distintos o una ventana privada para no cruzarlas.

Cómo autorizar cada proveedor: [guía paso a paso](providers/connect.es.md).
