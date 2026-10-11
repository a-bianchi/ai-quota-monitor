# Firmware para la Waveshare ESP32-C6-LCD-1.47

[English](README.md) · **Español**

Sketch de Arduino que lee `GET /usage` de tu propio servidor y muestra los cupos personales y laborales en la pantalla de la placa.

**Seguí la [guía del firmware](../../docs/firmware.es.md)**: explica, paso a paso, cómo instalar las herramientas y las bibliotecas, cargar la placa y resolver los problemas habituales.

Referencia rápida:

- Sketch a abrir, según cómo vayas a apoyar la placa: [`ai_quota_monitor_vertical/ai_quota_monitor_vertical.ino`](ai_quota_monitor_vertical/ai_quota_monitor_vertical.ino) (pantalla de 172 × 320) o [`ai_quota_monitor_horizontal/ai_quota_monitor_horizontal.ino`](ai_quota_monitor_horizontal/ai_quota_monitor_horizontal.ino) (320 × 172). Los dos muestran una cuenta por página, con letra grande, y van rotando. Solo difieren en `layout.h`.
- Configuración: en la carpeta del sketch elegido, copiá `config.example.h` a `config.h` y completá el Wi-Fi, la URL de tu API y su `DEVICE_TOKEN`. `config.h` está ignorado por Git.
- Placa: **ESP32C6 Dev Module**. Partition Scheme: **Huge APP (3MB No OTA/1MB SPIFFS)**.
- Versiones: esp32 by Espressif Systems 3.3.12, LVGL 8.3.10 del ZIP del demo de Waveshare, ArduinoJson 7.4.3.

Estado: funciona en la placa física por HTTP. Todavía no se probó una dirección `https://` en la placa.

Los drivers de pantalla están adaptados del demo oficial de Waveshare. Ver [`ai_quota_monitor_vertical/NOTICE.md`](ai_quota_monitor_vertical/NOTICE.md).
