# Firmware for the Waveshare ESP32-C6-LCD-1.47

**English** · [Español](README.es.md)

Arduino sketch that reads `GET /usage` from your own server and shows the personal and work quotas on the board's screen.

**Follow the [firmware guide](../../docs/firmware.md)**: it explains, step by step, how to install the tools and libraries, flash the board and solve the usual problems.

Quick reference:

- Sketch to open, depending on how the board will stand: [`ai_quota_monitor_vertical/ai_quota_monitor_vertical.ino`](ai_quota_monitor_vertical/ai_quota_monitor_vertical.ino) (172 × 320 screen) or [`ai_quota_monitor_horizontal/ai_quota_monitor_horizontal.ino`](ai_quota_monitor_horizontal/ai_quota_monitor_horizontal.ino) (320 × 172). Both show one account per page, in large text, and rotate. They only differ in `layout.h`.
- Configuration: in the folder of the sketch you chose, copy `config.example.h` to `config.h` and fill in Wi-Fi, the URL of your API and its `DEVICE_TOKEN`. `config.h` is ignored by Git.
- Board: **ESP32C6 Dev Module**. Partition Scheme: **Huge APP (3MB No OTA/1MB SPIFFS)**.
- Versions: esp32 by Espressif Systems 3.3.12, LVGL 8.3.10 from the Waveshare demo ZIP, ArduinoJson 7.4.3.

Status: it compiles and its parser and layout were checked on a computer against real API responses. It has **not** been tested on the physical board yet; only the official Waveshare demo was.

The screen drivers are adapted from the official Waveshare demo. See [`ai_quota_monitor_vertical/NOTICE.md`](ai_quota_monitor_vertical/NOTICE.md).
