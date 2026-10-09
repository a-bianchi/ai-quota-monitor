# API v1

[English](api.md) · **Español**

En [`docs/collections/`](collections/) hay una colección y un environment de Postman listos para importar con estos endpoints.

## Sesión del panel

La usa el [panel web](panel.es.md); es opcional. Se habilita definiendo `AUTH_USERNAME` y `AUTH_PASSWORD_HASH`; sin ellas, estos endpoints responden 503. `DEVICE_TOKEN` y `COLLECTOR_TOKEN` son independientes y siguen funcionando igual.

`POST /auth/login`, sin `Authorization`:

```json
{"username": "tu_usuario", "password": "tu_contraseña"}
```

Respuesta 200:

```json
{"tokenType": "Bearer", "accessToken": "aqm_a_...", "expiresIn": 900, "refreshToken": "aqm_r_..."}
```

- `POST /auth/refresh` con `{"refreshToken": "..."}` devuelve la misma forma con tokens nuevos. El refresh token usado deja de servir; presentarlo otra vez revoca la sesión.
- `POST /auth/logout` revoca la sesión del access token de `Authorization` o del `{"refreshToken": "..."}` del body. Responde 204.
- `accessToken` dura 15 minutos; `refreshToken` dura 30 días sin uso y solo sirve para `/auth/refresh` y `/auth/logout`.
- 401 si las credenciales son incorrectas o el refresh token es inválido, venció o ya fue usado; 400 si el JSON es inválido; 429 con `Retry-After` después de demasiados logins fallidos.

El servidor guarda solo hashes de los tokens, así las sesiones sobreviven a reinicios. Ver [qué rutas aceptan una sesión del panel](panel.es.md#sesión).

## Configuración desde el panel

Solo con sesión del panel; los tokens fijos no se aceptan.

- `GET /accounts` lista las cuentas configuradas y qué se puede configurar en cada una: `login` (`codex`, `claude` o vacío) y `credential` (`null`, o `{"configured": true, "origin": "stored" | "env"}`). `credentialStorage` indica si hay `MASTER_KEY`. Nunca se devuelven los valores de las credenciales.
- `PUT /accounts/{id}/credential` con `{"value": "..."}` guarda o reemplaza la credencial de una cuenta que usa `credentialEnv`. 404 si la cuenta no usa una, 400 si el valor está vacío o tiene varias líneas, 503 sin `MASTER_KEY`.

## Lectura

`GET /usage`, con `Authorization: Bearer DEVICE_TOKEN`. También acepta `COLLECTOR_TOKEN` y una sesión del panel. Filtro opcional `?scope=personal` o `?scope=work`. `GET /healthz` es público y solo indica que el proceso HTTP responde; no garantiza que los proveedores estén disponibles.

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
      "resetsAt": "2026-10-08T20:20:00-03:00",
      "usageLevel": "low"
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
- `usageLevel`: lo calcula el servidor con `used / limit`: `low` (menos de 50%), `medium` (de 50% a 80% inclusive), `high` (más de 80%), `unknown` (falta `used` o `limit`, o el límite es 0). Solo aparece en las respuestas; un reporte de ingestión que lo incluya se rechaza.
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

POST y GET `/accounts/{id}/codex/login`, con `Authorization: Bearer COLLECTOR_TOKEN` o una sesión del panel. Solo cuentas provider=openai, source=direct. POST inicia device-code login y devuelve código/URL oficiales; GET informa estado. DEVICE_TOKEN no tiene acceso. Si no hay COLLECTOR_TOKEN ni login del panel configurado, los endpoints están deshabilitados (503). Ver [flujo](providers/codex.es.md).

## Login Claude en homelab

`/accounts/{id}/claude/login`, con `Authorization: Bearer COLLECTOR_TOKEN` o una sesión del panel. Solo cuentas provider=anthropic, source=direct. POST inicia el login y devuelve `verificationUrl`; PUT con `{"code":"codigo#estado"}` lo completa; GET informa el estado. DEVICE_TOKEN no tiene acceso. Experimental y no permitido por los términos de Anthropic: ver [flujo y advertencia](providers/claude-subscription.es.md).
