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

## Placa usada y dónde comprarla

Este proyecto se desarrolló sobre una placa concreta y está basado en ella: la **Waveshare ESP32-C6-LCD-1.47**, comprada como un kit publicado como "ESP32-C6 con LCD 1.47", 172×320, LED RGB, Wi-Fi 6, tarjeta SD". Los pines, el driver de pantalla, las bibliotecas y la [guía del firmware](firmware.es.md) son para este modelo.

Dónde se puede comprar:

| Tienda | Enlace |
|---|---|
| Mercado Libre Argentina (donde se compró esta unidad) | <https://www.mercadolibre.com.ar/kit-esp32-c6-con-lcd-147--172320-led-rgb-wifi-6-tarjeta-sd/up/MLAU3223607467> |
| Amazon | <https://www.amazon.com/dp/B0DNZBXF8G> |
| AliExpress | <https://www.aliexpress.us/item/3256808808244278.html> |

Las publicaciones y los vendedores cambian, y estos enlaces pueden dejar de funcionar. Antes de comprar, verificá que la publicación sea de la **ESP32-C6-LCD-1.47**: pantalla de 1,47", 172 × 320, ESP32-C6 y **sin touch**. Las placas parecidas (ESP32-S3-LCD-1.47, las versiones Touch, otros tamaños de pantalla) usan otros pines o controladores, y este firmware no funciona en ellas tal como está.

## Conexiones y montaje

La pantalla, el LED RGB y la ranura microSD vienen integrados en la placa: no hay nada que cablear ni soldar. El único cable es el USB-C, que alimenta la placa y sirve para programarla.

Para programar la placa, seguí la [guía del firmware](firmware.es.md). No hay instrucciones de montaje validadas todavía: dependen de la [carcasa](../enclosure/README.es.md), que está pendiente.

## Estado y próximos pasos

- El demo oficial de Waveshare se compiló, se cargó y se vio en la placa real, con los pines listados arriba.
- El [firmware](firmware.es.md) del proyecto compila; todavía no se probó en una placa física.
- Próximo paso: cargar el firmware y revisar pantalla, Wi-Fi y legibilidad.
- La vista **Display** del [panel web](panel.es.md) muestra el mismo contenido en un navegador.

## Fuentes


- [Documentación Waveshare](https://docs.waveshare.com/ESP32-C6-LCD-1.47)
- [Producto del fabricante](https://www.waveshare.com/ESP32-C6-LCD-1.47.htm)
- Tiendas: ver [placa usada y dónde comprarla](#placa-usada-y-dónde-comprarla).
