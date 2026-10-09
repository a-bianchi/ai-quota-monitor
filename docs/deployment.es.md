# Despliegue

[English](deployment.md) · **Español**

## Docker Compose / homelab

Seguí el README. El volumen `usage-data` conserva las observaciones entre reinicios; `docker compose --env-file server/.env down -v` lo elimina. Los datos se guardan por combinación fuente/cuenta: cambiar de mock a real no reutiliza datos ficticios.

Para que el ESP32 acceda desde tu LAN, configurá `HOST_BIND` con la IP LAN del servidor o `0.0.0.0` y recreá el servicio. Usá firewall para limitar acceso a tu red. El valor por defecto solo permite conexiones locales.

Con Go local y con Docker se utiliza `CONFIG_NAME` y los archivos de `server/configs/`. El Dockerfile incluye las configuraciones, y Compose monta ese directorio para permitir cambios sin reconstruir la imagen. `.env` queda excluido de la imagen; los secretos se inyectan por entorno.

## Coolify o nube

Se puede desplegar `compose.yaml` desde Git, configurar las variables en el panel y conservar el volumen. Configurá el proxy de la plataforma hacia el puerto 8080 del contenedor; puede requerir quitar el mapeo `ports` de Compose. No se probó un despliegue real en Coolify todavía.

Para conexiones fuera de una red confiable, publicá HTTPS detrás de un proxy con certificado válido. Los tokens de lectura/ingestión siguen siendo obligatorios. No abrir el puerto HTTP directamente a Internet. El firmware deberá validar el certificado TLS.

El servidor no necesita modelos de IA ni GPU; consulta datos de consumo. Codex corre en el propio servidor. Su sesión se autoriza desde un navegador mediante código y se guarda en /data/codex. No depende de la Mac. El volumen contiene credenciales: respaldarlo y restringir acceso. La imagen incluye Node/Codex, por eso es mayor que una imagen de Go puro.

Persistencia: un único proceso con archivo JSON atómico. No escalar a varias réplicas sobre el mismo archivo; una futura implementación compartida necesitará otra estrategia. Respaldar el volumen si necesitás conservar datos. No almacena historial, solo la última observación por cuenta.
