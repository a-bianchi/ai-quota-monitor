# Roadmap

[English](roadmap.md) · **Español**

## Implementado

- Servidor Go sin dependencias externas.
- GET /usage con filtro personal/trabajo y Bearer token.
- Modo de ejemplo con seis cuentas.
- Cliente API OpenRouter con errores y timeout.
- Ingestión de observaciones para cuentas collector.
- Persistencia de última observación y detección de datos viejos.
- Docker Compose, documentación y CI de Go.
- Panel web incluido en el servidor: login con sesiones rotativas, tablero, vista previa de la pantalla del ESP32 y configuración de proveedores.
- Codex en el servidor: login HTTP por device code, perfiles aislados y polling automático.
- Clientes experimentales de Cursor y de la suscripción de Claude.
- Credenciales de proveedores cargadas desde el panel y guardadas cifradas.
- `usageLevel` por cupo en `GET /usage`.
- Nombres de cupos en español o inglés y fechas en la zona horaria pedida.
- Firmware para ESP32-C6 con Arduino IDE y LVGL 8.3: tiene guía paso a paso y funciona en la placa física.
- Carcasa de dos piezas para imprimir en 3D, con tapa a presión: modelo paramétrico, STL y 3MF; impresa.
- Despliegue en un servidor Coolify local.

## Próximo

1. Validar OpenRouter con una management key propia.
2. Validar login y cupos reales de ambas cuentas con el login HTTP implementado.
3. Probar el firmware por HTTPS con un certificado real, la reconexión Wi-Fi y ejecuciones largas.
4. Validar Cursor con una sesión real; implementar Kiro si vale la pena (ver [proveedores](providers/README.es.md)).
5. Validar Claude con una cuenta real; reemplazarlo si Anthropic publica una vía oficial.
6. Documentar el material y el perfil de impresión de la carcasa.

## Decisiones abiertas

- Frescura por cupo si hay observaciones independientes.
- Tokens individuales por recolector y permisos por cuenta.

Las integraciones experimentales todavía no se validaron con cuentas reales.
