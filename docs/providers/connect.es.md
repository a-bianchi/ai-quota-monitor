# Cómo conectar cada proveedor

[English](connect.md) · **Español**

Guía paso a paso. Cada proveedor se conecta una sola vez; después el servidor consulta solo cada `POLL_INTERVAL` (5 minutos por defecto).

| Proveedor | Cómo se conecta | Dónde queda la credencial | Estado |
|---|---|---|---|
| [OpenRouter](#openrouter) | Pegar una management key | `server/.env` o el panel | Implementado |
| [Codex (ChatGPT)](#codex-chatgpt) | Login por código desde el navegador | `DATA_DIR/codex/<cuenta>/` | Implementado; login real depende de cada cuenta |
| [Cursor](#cursor) | Pegar la cookie de sesión | `server/.env` o el panel | Experimental |
| [Claude](#claude-suscripción) | Login desde el navegador y pegar un código | `DATA_DIR/claude/<cuenta>/` | Experimental; no permitido por los términos de Anthropic |
| Kiro | — | — | Pendiente |

## Antes de empezar

1. Copiá el archivo de ejemplo: `cp server/.env.example server/.env`.
2. Generá dos tokens distintos con `openssl rand -hex 32` y pegalos en `DEVICE_TOKEN` y `COLLECTOR_TOKEN`. Son de esta API, no de los proveedores.
3. Poné `CONFIG_NAME=config.live.json` para usar cuentas reales en lugar de las de ejemplo.
4. Arrancá con `make run` (local) o `make docker-up` (Docker).

Los pasos de abajo usan `curl` y `server/.env`. Si habilitaste el [panel web](../panel.es.md), su vista **Configuración** hace lo mismo desde el navegador: inicia los logins de Codex y Claude y guarda las keys de OpenRouter y Cursor cifradas en el servidor (requiere `MASTER_KEY`).

Cada cuenta se declara en `server/configs/config.live.json`. Si no usás un proveedor, borrá su entrada. Después de cambiar `.env` o la configuración hay que reiniciar el servidor.

## Cómo verificar cualquier cuenta

```sh
curl -H 'Authorization: Bearer TU_COLLECTOR_TOKEN' http://localhost:8080/usage
```

Mirá el `status` de la cuenta:

| `status` | Qué significa | Qué hacer |
|---|---|---|
| `ok` | Datos recientes | Nada |
| `stale` | Hay datos, pero viejos o la última consulta falló | Esperar un ciclo; si sigue, revisar logs |
| `auth_required` | Falta la credencial, está mal o venció | Repetir los pasos de conexión de ese proveedor |
| `unavailable` | Nunca se obtuvieron datos | Revisar logs y conexión a internet del servidor |

La primera lectura puede tardar unos segundos después de arrancar.

## OpenRouter

Muestra créditos comprados y gastados, en USD.

1. Entrá a <https://openrouter.ai/settings/management-keys> y creá una **management key**. Una key común de inferencia no sirve para consultar créditos.
2. Pegala en `server/.env`:

   ```dotenv
   OPENROUTER_PERSONAL_KEY=tu_management_key
   ```

3. La cuenta en la configuración:

   ```json
   {"id":"openrouter-personal","name":"OpenRouter","provider":"openrouter","scope":"personal","source":"direct","credentialEnv":"OPENROUTER_PERSONAL_KEY"}
   ```

4. Reiniciá y verificá. Esperado: un cupo `balance` en `usd`.

La key no vence sola. Si la revocás, la cuenta pasa a `auth_required`.

## Codex (ChatGPT)

Muestra los límites de Codex de tu suscripción de ChatGPT. No son todos los límites del chat normal. Requiere Codex CLI (la imagen Docker ya lo incluye; en local se indica con `CODEX_BIN`).

No se pega ninguna credencial: iniciás sesión desde tu navegador y el servidor guarda la sesión.

1. La cuenta en la configuración (el provider se llama `openai`):

   ```json
   {"id":"codex-personal","name":"Codex personal","provider":"openai","scope":"personal","source":"direct"}
   ```

2. Pedí un código de login:

   ```sh
   curl -X POST http://localhost:8080/accounts/codex-personal/codex/login \
     -H 'Authorization: Bearer TU_COLLECTOR_TOKEN'
   ```

   Respuesta `202` con `verificationUrl` y `userCode`.

3. Abrí `verificationUrl`, iniciá sesión con la cuenta correcta, escribí el `userCode` y autorizá.
4. Consultá el estado con el mismo endpoint por `GET` hasta que diga `authenticated`.
5. Verificá con `GET /usage`.

Para tener personal y laboral, repetí los pasos con cada ID (`codex-personal`, `codex-work`). El ID no comprueba qué cuenta elegiste en el navegador: usá perfiles distintos o una ventana privada para no cruzarlas.

Si tu organización restringe el login por código de dispositivo, el paso 3 puede fallar; eso lo decide el administrador. Más detalle en [codex.es.md](codex.es.md).

## Cursor

Muestra el porcentaje usado del plan y la fecha de renovación. Usa el endpoint del dashboard de cursor.com, que no está documentado por Cursor y puede cambiar. No requiere ser administrador.

1. Abrí <https://cursor.com/dashboard> e iniciá sesión con la cuenta que querés medir.
2. Abrí las herramientas de desarrollador (`F12` o `Cmd+Option+I`).
3. Andá a **Application** (en Firefox, **Storage**) → **Cookies** → `https://cursor.com`.
4. Buscá `WorkosCursorSessionToken` y copiá el **valor completo**.
5. Pegalo en `server/.env`:

   ```dotenv
   CURSOR_WORK_SESSION=valor_copiado
   ```

6. La cuenta en la configuración:

   ```json
   {"id":"cursor-work","name":"Cursor trabajo","provider":"cursor","scope":"work","source":"direct","credentialEnv":"CURSOR_WORK_SESSION"}
   ```

7. Reiniciá y verificá. Esperado: un cupo `included` en `percent`, y según el plan también `auto`, `api`, `on-demand` o `team-on-demand`.

A tener en cuenta:

- La cookie da acceso a tu cuenta de Cursor: tratala como una contraseña.
- Vence y no se renueva sola. Cuando la cuenta pase a `auth_required`, repetí los pasos 1 a 5 y reiniciá.
- Cerrar sesión en cursor.com puede invalidarla antes.
- Algunas cuentas Enterprise no exponen ningún medidor con límite; en ese caso queda `unavailable`.

Para una segunda cuenta, agregá otra entrada con otra variable (por ejemplo `CURSOR_PERSONAL_SESSION`) y, si usás Docker, sumala también en `compose.yaml`. Más detalle en [cursor.es.md](cursor.es.md).

## Claude (suscripción)

Muestra las ventanas de 5 horas y 7 días de Claude Pro/Max.

**Antes de usarlo:** los términos de Anthropic no permiten que una aplicación de terceros inicie sesión con Claude.ai ni guarde sus tokens, y eso es lo que hace este provider. El riesgo es sobre tu cuenta. Leé la [advertencia completa](claude-subscription.es.md).

1. La cuenta en la configuración:

   ```json
   {"id":"claude-personal","name":"Claude personal","provider":"anthropic","scope":"personal","source":"direct"}
   ```

2. Pedí la URL de login:

   ```sh
   curl -X POST http://localhost:8080/accounts/claude-personal/claude/login \
     -H 'Authorization: Bearer TU_COLLECTOR_TOKEN'
   ```

3. Abrí `verificationUrl`, iniciá sesión con la cuenta correcta y autorizá. Copiá el código que muestra la página (`codigo#estado`).
4. Envialo antes de 15 minutos:

   ```sh
   curl -X PUT http://localhost:8080/accounts/claude-personal/claude/login \
     -H 'Authorization: Bearer TU_COLLECTOR_TOKEN' \
     -H 'Content-Type: application/json' \
     -d '{"code":"codigo#estado"}'
   ```

5. Verificá con `GET /usage`. Esperado: cupos `five-hour` y `seven-day` en `percent`.

El servidor renueva la sesión solo. Si pasa a `auth_required`, repetí los pasos 2 a 4. Más detalle en [claude-subscription.es.md](claude-subscription.es.md).

## Seguridad

- `server/.env` y `server/data/` están en `.gitignore`. No los subas ni los pegues en issues.
- En Docker, las sesiones de Codex y Claude viven en el volumen `usage-data`; sus backups también son secretos.
- Usá `DEVICE_TOKEN` en el display y `COLLECTOR_TOKEN` solo para administrar.
