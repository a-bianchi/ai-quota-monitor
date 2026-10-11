# Hardware

**English** · [Español](hardware.es.md)

Target model: **Waveshare ESP32-C6-LCD-1.47** (not the S3 or the Touch version). 172 × 320 display, ST7789 controller over SPI, 2.4 GHz Wi-Fi, RGB LED and microSD. USB powered; no battery, charger, IMU or touch are assumed.

Initial bill of materials:

- Waveshare ESP32-C6-LCD-1.47 board.
- USB-C data cable for flashing and power.
- USB port or compatible power supply.
- Printed [enclosure](../enclosure/README.md), optional.

MicroSD is not required for the MVP.

According to the manufacturer's documentation, LCD: MOSI GPIO6, SCLK GPIO7, CS GPIO14, DC GPIO15, RST GPIO21, backlight GPIO22; RGB GPIO8. Confirm against the physical revision before wiring or flashing. The manufacturer recommends 50% brightness or lower for prolonged use.

## Board used and where to buy it

This project was developed on, and is based on, one specific board: the **Waveshare ESP32-C6-LCD-1.47**, bought as a kit sold as "ESP32-C6 with 1.47" LCD, 172×320, RGB LED, Wi-Fi 6, SD card". The pins, the screen driver, the libraries and the [firmware guide](firmware.md) are all for this model.

Where it can be bought:

| Store | Link |
|---|---|
| Mercado Libre Argentina (where this unit was bought) | <https://www.mercadolibre.com.ar/kit-esp32-c6-con-lcd-147--172320-led-rgb-wifi-6-tarjeta-sd/up/MLAU3223607467> |
| Amazon | <https://www.amazon.com/dp/B0DNZBXF8G> |
| AliExpress | <https://www.aliexpress.us/item/3256808808244278.html> |

Listings and sellers change and these links may stop working. Before buying, check that the listing is for the **ESP32-C6-LCD-1.47**: 1.47" screen, 172 × 320, ESP32-C6, **without touch**. Similar boards (ESP32-S3-LCD-1.47, the Touch versions, other screen sizes) use other pins or controllers and this firmware will not work on them as is.

## Connections and assembly

The screen, the RGB LED and the microSD slot are built into the board: there is nothing to wire or solder. The only cable is the USB-C one, which powers the board and is used to flash it.

To program the board, follow the [firmware guide](firmware.md). To mount it, see the [enclosure](../enclosure/README.md).

## Status and next steps

- The official Waveshare demo was compiled, uploaded and shown on the real board, with the pins listed above.
- The project [firmware](firmware.md) runs on the physical board.
- The [enclosure](../enclosure/README.md) is designed and printed.
- Next step: test HTTPS with a real certificate.
- The **Display** view of the [web panel](panel.md) previews the same content in a browser.

## Sources


- [Waveshare documentation](https://docs.waveshare.com/ESP32-C6-LCD-1.47)
- [Manufacturer product page](https://www.waveshare.com/ESP32-C6-LCD-1.47.htm)
- Stores: see [board used and where to buy it](#board-used-and-where-to-buy-it).
