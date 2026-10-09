# API v1

[English](api.en.md) · **Español**

## Login con usuario y contraseña

Opcional. Se habilita definiendo `AUTH_USERNAME` y `AUTH_PASSWORD` (8+ caracteres) junto con `COLLECTOR_TOKEN`; sin ellas, estos endpoints responden 503. `DEVICE_TOKEN` y `COLLECTOR_TOKEN` siguen funcionando igual.

`POST /auth/login`, sin `Authorization`:

```json
{"username": "tu_usuario", "password": "tu_contraseña"}
```

Respuesta 200:

```json
{"tokenType": "Bearer", "accessToken": "<jwt>", "expiresIn": 900, "refreshToken": "<jwt>"}
```

`POST /auth/refresh` con `{"refreshToken": "<jwt>"}` devuelve la misma forma, con tokens nuevos.

- `accessToken` dura 15 minutos y se usa como `Authorization: Bearer <accessToken>`. Da los mismos permisos que `COLLECTOR_TOKEN`: leer `/usage`, ingresar datos e iniciar logins de proveedores.
- `refreshToken` dura 30 días y solo sirve para `/auth/refresh`.
- 401 si las credenciales o el refresh token son inválidos o vencieron; 400 si el JSON es inválido.

Los tokens son JWT firmados (HS256) y el servidor no los guarda: siguen valiendo después de reiniciar. No hay cierre de sesión ni revocación individual; para invalidar todos, cambiá `AUTH_PASSWORD` o `COLLECTOR_TOKEN`. Un login fallido demora un segundo en responder; no hay bloqueo por intentos, así que usá una contraseña fuerte y HTTPS si la API sale de tu red.

## Lectura

`GET /usage`, con `Authorization: Bearer DEVICE_TOKEN`. También acepta `COLLECTOR_TOKEN`, para poder probar todo con un solo token desde Postman. Filtro opcional `?scope=personal` o `?scope=work`. `GET /healthz` es público y solo indica que el proceso HTTP responde; no garantiza que los proveedores estén disponibles.

Encabezados opcionales:

| Encabezado | Valores | Sin encabezado |
|---|---|---|
| `X-Timezone` | Zona IANA, por ejemplo `Europe/Madrid` o `UTC` | `America/Argentina/Buenos_Aires` (UTC-3) |
| `Accept-Language` | `es` o `en` | `es` |

`X-Timezone` convierte `generatedAt`, `updatedAt` y `resetsAt` a esa zona; el instante es el mismo, cambia el desplazamiento (`-03:00`, `+02:00`, o `Z` para UTC). Una zona desconocida devuelve 400. `Accept-Language` traduce `quotas[].name`; se usa el primer idioma soportado de la lista y cualquier otro cae en español. Los `id` de los cupos nunca se traducen. Los nombres de cuenta vienen de la configuración y no se traducen, igual que los nombres propios que envíe un recolector.

Respuesta:

```json
{
  "schemaVersion": 1,
  "generatedAt": "2026-10-08T17:41:00-03:00",
  "accounts": [{
    "id": "codex-work",
    "name": "Codex trabajo",
    "provider": "openai",
    "scope": "work",
    "source": "collector",
    "updatedAt": "2026-10-08T17:40:00-03:00",
    "status": "ok",
    "quotas": [{
      "id": "5h",
      "name": "5 horas",
      "unit": "percent",
      "used": 11,
      "limit": 100,
      "resetsAt": "2026-10-08T20:20:00-03:00"
    }]
  }]
}
```

Los valores anteriores son ficticios. Fechas RFC3339/ISO 8601, en hora de Argentina (UTC-3) salvo que se pida otra zona con `X-Timezone`. `updatedAt` es la observación del proveedor; `generatedAt` es la generación de esta respuesta. No usar `generatedAt` para inferir frescura.

- `provider`: `openai`, `cursor`, `kiro`, `openrouter`, `anthropic`.
- `scope`: `personal`, `work`.
- `source`: `mock`, `direct`, `collector`.
- `status`: `ok`, `stale`, `auth_required`, `unavailable`.
- `unit`: `percent`, `credits`, `usd`, `tokens`.
- `used`, `limit`, `updatedAt`, `resetsAt` pueden ser `null` cuando corresponda.
- `quotas` es una lista, vacía cuando todavía no hay datos.

Saldo = `limit - used` si ambos existen. Si el consumo supera el límite, mostrar agotado y conservar los valores originales. En porcentajes el límite es 100. No sumar porcentajes de cupos distintos. OpenRouter devuelve créditos acumulados comprados y usados: mostrar saldo USD; ese total no es un cupo mensual ni una barra de suscripción.

`stale`: pasó `STALE_AFTER`, una ventana ya se renovó o falló una actualización y existen datos anteriores. `auth_required`: clave ausente/rechazada. `unavailable`: sin observación utilizable o error de almacenamiento. Puede haber valores anteriores aun con error: la pantalla debe mostrar su estado.

El MVP tiene un timestamp/estado por cuenta. Un recolector debe enviar juntos los cupos de una misma observación; datos independientes requerirán extender el contrato.

## Ingestión

`PUT /accounts/{id}/usage` con `Authorization: Bearer COLLECTOR_TOKEN`:

```json
{
  "updatedAt": "2026-10-08T20:40:00Z",
  "quotas": [{"id":"5h","name":"5 horas","unit":"percent","used":11,"limit":100,"resetsAt":null}]
}
```

Solo admite cuentas declaradas con `source: collector`; metadatos de identidad provienen de la configuración, no del payload. Máximo 64 KiB y 16 cupos, IDs únicos y valores finitos no negativos. Rechaza observaciones anteriores al último dato y timestamps futuros con más de un minuto de margen. La lista reemplaza todos los cupos de la cuenta. Respuesta 200 al persistir, 400 payload inválido, 401 token incorrecto, 404 cuenta no ingerible, 409 observación anterior y 503 si la ingestión está deshabilitada.

`COLLECTOR_TOKEN` es compartido por los recolectores de este MVP: puede actualizar todas las cuentas collector. También permite leer `GET /usage`. No usarlo como token del dispositivo. No envíes credenciales de proveedor en el reporte.

## Login Codex en homelab

POST y GET `/accounts/{id}/codex/login`, con `Authorization: Bearer COLLECTOR_TOKEN`. Solo cuentas provider=openai, source=direct. POST inicia device-code login y devuelve código/URL oficiales; GET informa estado. DEVICE_TOKEN no tiene acceso. Si COLLECTOR_TOKEN está vacío, los endpoints están deshabilitados (503). Ver [flujo](providers/codex.md).

## Login Claude en homelab

`/accounts/{id}/claude/login`, con `Authorization: Bearer COLLECTOR_TOKEN`. Solo cuentas provider=anthropic, source=direct. POST inicia el login y devuelve `verificationUrl`; PUT con `{"code":"codigo#estado"}` lo completa; GET informa el estado. DEVICE_TOKEN no tiene acceso. Experimental y no permitido por los términos de Anthropic: ver [flujo y advertencia](providers/claude-subscription.md).
