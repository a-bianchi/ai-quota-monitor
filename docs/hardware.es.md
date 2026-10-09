# Hardware

[English](hardware.md) · **Español**

Modelo objetivo: **Waveshare ESP32-C6-LCD-1.47** (no S3 ni versión Touch). Pantalla 172 × 320, controlador ST7789 por SPI, Wi-Fi de 2,4 GHz, LED RGB y microSD. Alimentación USB; no se asumen batería, cargador, IMU ni táctil.

Lista de materiales inicial:

- Placa Waveshare ESP32-C6-LCD-1.47.
- Cable USB-C de datos para programar y alimentar.
- Puerto USB o fuente compatible.
- Carcasa impresa, pendiente de diseño y validación.

MicroSD no requerida para el MVP.

Según la documentación del fabricante, LCD: MOSI GPIO6, SCLK GPIO7, CS GPIO14, DC GPIO15, RST GPIO21, backlight GPIO22; RGB GPIO8. Confirmar contra la revisión física antes de cablear/programar. El fabricante recomienda brillo de 50% o menor para uso prolongado.

Fuentes:

- [Documentación Waveshare](https://docs.waveshare.com/ESP32-C6-LCD-1.47)
- [Producto del fabricante](https://www.waveshare.com/ESP32-C6-LCD-1.47.htm)
- [Referencia de compra en Argentina](https://www.mercadolibre.com.ar/kit-esp32-c6-con-lcd-147--172320-led-rgb-wifi-6-tarjeta-sd/up/MLAU3223607467)

La referencia de compra puede cambiar; verificá el modelo en lugar de depender del vendedor. Todavía no se probó el firmware en una placa física.
