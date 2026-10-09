# Codex en el servidor / homelab

La API Go inicia el login por código de dispositivo mediante Codex App Server. El usuario autoriza en su navegador y Codex guarda/renueva las credenciales en el servidor. No se requiere un recolector en la Mac ni copiar cookies o access tokens a la API.

## Configuración

En `server/configs/config.live.json`, cada cuenta Codex usa:

```json
{"id":"codex-work","name":"Codex trabajo","provider":"openai","scope":"work","source":"direct"}
```

En `server/.env`:

```dotenv
CONFIG_NAME=config.live.json
COLLECTOR_TOKEN=un_token_aleatorio_de_al_menos_32_caracteres
```

Generar el token con `openssl rand -hex 32`. Debe ser distinto de DEVICE_TOKEN. Reutilizá el COLLECTOR_TOKEN existente: es común a todos los proveedores y permite ingresar datos e iniciar/consultar logins; no es una credencial OpenAI. No usar el marcador literal. Reiniciar el servidor al cambiar configuración/tokens.

- Go local: `make run`; requiere Codex CLI, configurable con CODEX_BIN.
- Homelab/Docker: `make docker-up`; la imagen incluye Codex 0.162.0. Compose usa CODEX_BIN=codex, sin usar la ruta del ejecutable de tu Mac.
- LAN: ajustar HOST_BIND según deployment.md.

## Login HTTP

```sh
curl -X POST http://TU_SERVIDOR:8080/accounts/codex-work/codex/login \
  -H 'Authorization: Bearer TU_COLLECTOR_TOKEN'
```

Respuesta 202 de ejemplo:

```json
{
  "status":"pending",
  "verificationUrl":"https://auth.openai.com/codex/device",
  "userCode":"ABCD-1234",
  "expiresAt":"2026-10-09T16:00:00Z"
}
```

Abrir verificationUrl en tu navegador, introducir el código y elegir la cuenta/workspace del trabajo. Para la personal usar el endpoint de codex-personal. Es tu acceso normal; las restricciones de la organización siguen aplicando. Puede ser necesario habilitar device-code login en la configuración de seguridad de ChatGPT o pedir al administrador que habilite el método si está restringido.

El código de ejemplo no es válido. expiresAt es el límite local de espera (15 minutos); OpenAI puede invalidar el código antes. No compartir userCode ni la respuesta de login en un issue.

Estado:

```sh
curl http://TU_SERVIDOR:8080/accounts/codex-work/codex/login \
  -H 'Authorization: Bearer TU_COLLECTOR_TOKEN'
```

Estados: starting, pending, authenticated, auth_required, failed, unknown. Después de reiniciar, el estado se recupera al consultar la sesión. unknown no significa que la sesión se haya perdido. POST repetido mientras está pending devuelve el mismo intento, sin lanzar otro login. Si una lectura está en curso, puede responder 409; reintentar luego.

Cuando se completa el login, el servidor consulta inmediatamente los cupos. Luego actualiza cada POLL_INTERVAL (5m por defecto). GET /usage acepta DEVICE_TOKEN o COLLECTOR_TOKEN; DEVICE_TOKEN no permite iniciar logins.

## Persistencia y límites

Sesiones en DATA_DIR/codex/<account-id>/; en Docker son /data/codex/<account-id>/ dentro del volumen usage-data. Cada cuenta tiene credenciales separadas. Reiniciar conserva sesiones; borrar el volumen las elimina. Tratar el volumen y sus backups como secretos. No exponer esta carpeta ni ponerla en Git.

Solo se usan endpoints de cuenta: initialize, account/login/start (chatgptDeviceCode), account/read y account/rateLimits/read. No hay prompts, herramientas ni consumo de reset credits. La API no recibe ni devuelve cookies ni access/refresh tokens.

El ID codex-work no verifica qué cuenta elegiste: revisar cuenta/workspace durante el login. Se conserva la distinción de buckets reales y sus duraciones; no se inventan ventanas que no aparezcan en la respuesta. Estos límites corresponden a Codex/Work, no a todos los chats de ChatGPT.

Para uso personal autohospedado. La documentación de OpenAI distingue las integraciones locales/open source de servicios comerciales/hosted con autenticación App Server; no convertir estos endpoints en un servicio público multiusuario. Ver los requisitos actuales antes de ofrecerlo como servicio.

[App Server oficial](https://learn.chatgpt.com/docs/app-server) · [Autenticación](https://learn.chatgpt.com/docs/auth)

El flujo se verifica con un App Server simulado; los logins y políticas reales dependen de cada cuenta. Docker debe poder acceder por HTTPS a OpenAI.

## Recolector local anterior

El binario cmd/collector permanece como opción para configuraciones source=collector. No se necesita para config.live.json ni para el despliegue homelab actual.
