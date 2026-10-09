# Panel web

[English](panel.md) · **Español**

Un panel web chico servido por la misma API Go en `/`. Es HTML, CSS y Alpine.js incluidos dentro del binario del servidor: sin Node, sin paso de build y sin proceso aparte. Funciona igual con `make run` y con Docker Compose.

## Vistas

- **Servicios**: cuentas personales y laborales, con proveedor, estado, cupos y última actualización.
- **Display**: vista previa de la pantalla del ESP32 de 172 × 320 px con los datos reales de la API, alternando entre personal y laboral.
- **Configuración**: login de Codex (código y enlace de autorización a la vista), login de Claude, y carga o reemplazo de la key de OpenRouter y la cookie de Cursor.

Español e inglés, con selector; la elección se guarda en el navegador. El idioma inicial es español.

## Crear el usuario

No hay credenciales por defecto. El panel queda deshabilitado hasta que definas un usuario.

1. Generá el hash de la contraseña. La pide dos veces y no la muestra:

   ```sh
   make panel-password
   ```

   Imprime una línea, `AUTH_PASSWORD_HASH=pbkdf2-sha256:...`. La contraseña necesita al menos 6 caracteres.

2. Pegá esa línea en `server/.env` y elegí el usuario:

   ```dotenv
   AUTH_USERNAME=tu_usuario
   AUTH_PASSWORD_HASH=pbkdf2-sha256:600000:...
   ```

3. Opcional, para guardar credenciales de proveedores desde el panel, agregá una clave maestra:

   ```sh
   openssl rand -base64 32
   ```

   ```dotenv
   MASTER_KEY=el_valor_generado
   ```

4. Reiniciá (`make run` o `make docker-up`) y abrí <http://localhost:8080/>.

Con Docker y sin Go local, generá el hash dentro de la imagen:

```sh
docker compose --env-file server/.env run --rm -it api hash-password
```

Para cambiar la contraseña, generá un hash nuevo, reemplazá la línea y reiniciá. `AUTH_PASSWORD` en texto plano ya no se acepta: el servidor no arranca si está definida.

## Sesión

- `POST /auth/login` devuelve un access token (15 minutos) y un refresh token (30 días sin uso). El panel guarda ambos en `localStorage` y envía `Authorization: Bearer <accessToken>`. No se usan cookies.
- El panel renueva el access token un poco antes de que venza. Si una petición recibe 401, renueva y la repite una sola vez; si la renovación falla, limpia la sesión y vuelve al login.
- Cada renovación rota el refresh token e invalida el anterior. Usar un refresh token ya rotado revoca esa sesión.
- El servidor guarda solo hashes SHA-256 de los tokens, en `DATA_DIR/panel-sessions.json`, así las sesiones sobreviven a reinicios. Se conservan hasta 32 sesiones.
- `POST /auth/logout` revoca la sesión en el servidor y el panel borra los tokens del navegador.
- Los intentos de login están limitados: 5 fallos por dirección de cliente y 50 en total cada 15 minutos; después el servidor responde 429.
- La contraseña se guarda solo como hash PBKDF2-HMAC-SHA256 con 600.000 iteraciones y salt aleatorio.

Los tokens del panel son independientes de `DEVICE_TOKEN` y `COLLECTOR_TOKEN`, que siguen funcionando como antes para el ESP32 y los recolectores. El panel nunca guarda ni necesita esos dos tokens.

| Ruta | Sesión del panel | `DEVICE_TOKEN` | `COLLECTOR_TOKEN` |
|---|---|---|---|
| `GET /usage` | sí | sí | sí |
| `GET /accounts`, `PUT /accounts/{id}/credential` | sí | no | no |
| Rutas de login de Codex y Claude | sí | no | sí |
| `PUT /accounts/{id}/usage` (ingestión) | no | no | sí |

Los archivos HTML, CSS y JavaScript son públicos, como cualquier página de login; no contienen datos. Toda ruta que devuelve o cambia datos verifica la sesión en el servidor.

## Credenciales de proveedores

- Se guardan en el servidor en `DATA_DIR/credentials.json`, cifradas con AES-256-GCM usando una clave derivada de `MASTER_KEY`.
- Nunca se devuelven al navegador: el panel solo muestra si hay una configurada y de dónde viene.
- **Prioridad**: gana el valor guardado desde el panel; si no hay, se usa como respaldo la variable de entorno indicada en `credentialEnv` (por ejemplo `OPENROUTER_PERSONAL_KEY`).
- Sin `MASTER_KEY` el panel no puede guardar credenciales; las variables de entorno siguen funcionando.
- Si `MASTER_KEY` cambia o se pierde, las credenciales guardadas no se pueden leer y hay que cargarlas de nuevo.
- Codex y Claude conservan sus propias sesiones por cuenta en el servidor; no son API keys.

## Notas de seguridad

- `localStorage` puede ser leído por JavaScript que corra en la página; por eso el panel no carga scripts de terceros, nunca inserta HTML a partir de datos y se sirve con una Content Security Policy estricta (`script-src 'self'`, sin código inline ni `eval`). Alpine.js está incluido localmente con versión fija, en su build CSP; ver `server/internal/web/static/vendor/README.md`.
- Usá HTTPS cuando el panel se abra desde fuera de una red confiable: sin eso, la contraseña y los tokens viajan en claro. Ver [despliegue](deployment.es.md).
- Detrás de un proxy inverso todos los clientes comparten una dirección, así que el límite de login por dirección se comporta como uno global.
- Las sesiones, las credenciales y las sesiones de proveedores viven en el volumen de datos: tratá ese volumen y sus backups como secretos.

## Qué está verificado

Los tests automáticos cubren login, rotación, detección de reuso, logout, vencimiento, límite de intentos, protección de rutas, cifrado y prioridad de credenciales, y los archivos estáticos. La interfaz se probó de punta a punta contra un servidor corriendo con cuentas de ejemplo. No está verificado con cuentas reales desde el panel: un login completo de Codex o Claude y la lectura de OpenRouter con una key guardada desde el panel.
