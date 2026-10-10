# Third-party code

`Display_ST7789.h`, `Display_ST7789.cpp`, `LVGL_Driver.h` and `LVGL_Driver.cpp` are adapted from the official Waveshare demo for the ESP32-C6-LCD-1.47, example `Arduino/examples/LVGL_Arduino` inside:

https://files.waveshare.com/wiki/ESP32-C6-LCD-1.47/ESP32-C6-LCD-1.47-Demo.zip

The pins, the SPI settings and the ST7789 initialisation sequence are the manufacturer's. Changes: SD card, RGB, wireless test and LVGL demo code removed; the dummy touch input removed; pixel data sent with `SPI.writeBytes`; orientation chosen in `layout.h`. The example sources in that ZIP carry no license header or license file; they are credited here to Waveshare.

Libraries are not copied into this repository. They are installed separately and keep their own licenses:

- LVGL 8.3.10, MIT. Taken from `Arduino/libraries/lvgl` in the same ZIP, which includes `src/lv_conf.h` configured for this board.
- ArduinoJson 7.4.3, MIT. https://arduinojson.org
- Arduino core for ESP32 3.3.12 by Espressif Systems, LGPL-2.1 with Apache-2.0 components.
