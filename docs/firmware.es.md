# Firmware y pantalla — pendiente

[English](firmware.md) · **Español**

Destino: `firmware/esp32-c6-lcd-1.47/`. No hay firmware compilable todavía. Elegir y verificar toolchain/librerías compatibles con ESP32-C6 y ST7789 antes de implementar; usar los ejemplos oficiales Waveshare como referencia.

Configuración prevista: SSID, contraseña Wi-Fi, URL de API, token de lectura y zona horaria. Mantener esos valores fuera del repo. Las credenciales del proveedor no deben ir en el ESP32.

Diseño:

- PERSONAL en celeste y TRABAJO en violeta.
- Rotación automática cada 10s; sin botones táctiles ficticios.
- Una tarjeta por servicio, con saldo restante, cupos separados y renovación.
- OpenRouter en USD, sin barra mensual ni renovación inventada.
- Última actualización visible, distinto de la hora de lectura.
- Indicar MOCK cuando `source=mock`.
- Manejar `null`, stale, auth_required, unavailable y falta de Wi-Fi.
- LED RGB opcional para saldo bajo; umbrales configurables.

Validar legibilidad real en 172 × 320: simplificar o paginar si tres tarjetas resultan densas. Mostrar renovación por ventana o indicar claramente a cuál corresponde. El firmware debe validar TLS en conexiones HTTPS y mantener refresco HTTP separado del refresco gráfico.
