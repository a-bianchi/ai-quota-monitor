# Roadmap

## Implementado

- Servidor Go sin dependencias externas.
- GET /usage con filtro personal/trabajo y Bearer token.
- Modo de ejemplo con seis cuentas.
- Cliente API OpenRouter con errores y timeout.
- Ingestión de observaciones para cuentas collector.
- Persistencia de última observación y detección de datos viejos.
- Docker Compose, documentación y CI de Go.
- Codex en el servidor: login HTTP por device code, perfiles aislados y polling automático.

## Próximo

1. Validar OpenRouter con una management key propia.
2. Validar login y cupos reales de ambas cuentas con el login HTTP implementado.
3. Implementar firmware y validar LCD/Wi-Fi/TLS en la placa real.
4. Validar Cursor con una sesión real; evaluar acceso sin admin en Kiro.
5. Validar Claude con una cuenta real; reemplazarlo si Anthropic publica una vía oficial.
6. Diseñar, imprimir y verificar carcasa.

## Decisiones abiertas

- Toolchain del firmware y librería de UI.
- Frescura por cupo si hay observaciones independientes.
- Tokens individuales por recolector y permisos por cuenta.
- Modelo paramétrico de carcasa (OpenSCAD o CAD equivalente).
- Dashboard web para previsualizar la pantalla.

El backend todavía no tuvo despliegue real ni pruebas con cuentas reales. No hay instrucciones de montaje físico validadas.
