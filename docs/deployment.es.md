# Despliegue

[English](deployment.md) · **Español**

## Docker Compose / homelab

Seguí el README. El volumen `usage-data` conserva las observaciones entre reinicios; `docker compose --env-file server/.env down -v` lo elimina. Los datos se guardan por combinación fuente/cuenta: cambiar de mock a real no reutiliza datos ficticios.

Para que el ESP32 acceda desde tu LAN, configurá `HOST_BIND` con la IP LAN del servidor o `0.0.0.0` y recreá el servicio. Usá firewall para limitar acceso a tu red. El valor por defecto solo permite conexiones locales.

Con Go local y con Docker se utiliza `CONFIG_NAME` y los archivos de `server/configs/`. El Dockerfile incluye las configuraciones, y Compose monta ese directorio para permitir cambios sin reconstruir la imagen. `.env` queda excluido de la imagen; los secretos se inyectan por entorno.

## Panel web

El panel es parte del mismo binario y del mismo contenedor: se sirve en `/`, en el mismo puerto que la API, sin nada más que desplegar. Definí `AUTH_USERNAME`, `AUTH_PASSWORD_HASH` y `MASTER_KEY` en `server/.env` (o en las variables de la plataforma); Compose las pasa al contenedor. Ver [panel web](panel.es.md) para crear el usuario.

Las sesiones del panel (`panel-sessions.json`) y las credenciales de proveedores cifradas (`credentials.json`) viven en el volumen `usage-data`, junto a las sesiones de Codex y Claude. Guardá `MASTER_KEY` fuera de ese volumen: una copia del volumen sola no tiene que alcanzar para leer las credenciales. Perder `MASTER_KEY` implica cargarlas de nuevo.

Entrá al panel por HTTPS siempre que no esté en una red confiable. `HOST_BIND` controla quién puede conectarse, igual que para la API.

## Coolify o nube

Se puede desplegar `compose.yaml` desde Git, configurar las variables en el panel y conservar el volumen. Configurá el proxy de la plataforma hacia el puerto 8080 del contenedor; puede requerir quitar el mapeo `ports` de Compose. No se probó un despliegue real en Coolify todavía.

Para conexiones fuera de una red confiable, publicá HTTPS detrás de un proxy con certificado válido. Los tokens de lectura/ingestión siguen siendo obligatorios. No abrir el puerto HTTP directamente a Internet. El firmware deberá validar el certificado TLS.

El servidor no necesita modelos de IA ni GPU; consulta datos de consumo. Codex corre en el propio servidor. Su sesión se autoriza desde un navegador mediante código y se guarda en /data/codex. No depende de la Mac. El volumen contiene credenciales: respaldarlo y restringir acceso. La imagen incluye Node/Codex, por eso es mayor que una imagen de Go puro.

Persistencia: un único proceso con archivo JSON atómico. No escalar a varias réplicas sobre el mismo archivo; una futura implementación compartida necesitará otra estrategia. Respaldar el volumen si necesitás conservar datos. No almacena historial, solo la última observación por cuenta.
