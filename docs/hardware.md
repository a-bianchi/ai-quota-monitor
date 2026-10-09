# Hardware

**English** · [Español](hardware.es.md)

Target model: **Waveshare ESP32-C6-LCD-1.47** (not the S3 or the Touch version). 172 × 320 display, ST7789 controller over SPI, 2.4 GHz Wi-Fi, RGB LED and microSD. USB powered; no battery, charger, IMU or touch are assumed.

Initial bill of materials:

- Waveshare ESP32-C6-LCD-1.47 board.
- USB-C data cable for flashing and power.
- USB port or compatible power supply.
- Printed enclosure, pending design and validation.

MicroSD is not required for the MVP.

According to the manufacturer's documentation, LCD: MOSI GPIO6, SCLK GPIO7, CS GPIO14, DC GPIO15, RST GPIO21, backlight GPIO22; RGB GPIO8. Confirm against the physical revision before wiring or flashing. The manufacturer recommends 50% brightness or lower for prolonged use.

## Connections and assembly

The screen, the RGB LED and the microSD slot are built into the board: there is nothing to wire or solder. The only cable is the USB-C one, which powers the board and is used to flash it.

There are no validated assembly instructions yet. They depend on the [firmware](firmware.md) and the [enclosure](../enclosure/README.md), which are pending.

## Status and next steps

- The board is chosen; the firmware has not been tested on a physical board.
- Next step: check the pins against the real board and light up the screen with an official Waveshare example.
- Meanwhile, the **Display** view of the [web panel](panel.md) shows the content planned for the screen.

## Sources


- [Waveshare documentation](https://docs.waveshare.com/ESP32-C6-LCD-1.47)
- [Manufacturer product page](https://www.waveshare.com/ESP32-C6-LCD-1.47.htm)
- [Purchase reference in Argentina](https://www.mercadolibre.com.ar/kit-esp32-c6-con-lcd-147--172320-led-rgb-wifi-6-tarjeta-sd/up/MLAU3223607467)

The purchase reference may change; check the model instead of relying on the seller. The firmware has not been tested on a physical board yet.
