# Firmware para la Waveshare ESP32-C6-LCD-1.47

[English](README.md) · **Español**

Sketch de Arduino que lee `GET /usage` de tu propio servidor y muestra los cupos personales y laborales en la pantalla de la placa.

**Seguí la [guía del firmware](../../docs/firmware.es.md)**: explica, paso a paso, cómo instalar las herramientas y las bibliotecas, cargar la placa y resolver los problemas habituales.

Referencia rápida:

- Sketch a abrir: [`ai_quota_monitor/ai_quota_monitor.ino`](ai_quota_monitor/ai_quota_monitor.ino).
- Configuración: copiá `ai_quota_monitor/config.example.h` a `ai_quota_monitor/config.h` y completá el Wi-Fi, la URL de tu API y su `DEVICE_TOKEN`. `config.h` está ignorado por Git.
- Placa: **ESP32C6 Dev Module**. Partition Scheme: **Huge APP (3MB No OTA/1MB SPIFFS)**.
- Versiones: esp32 by Espressif Systems 3.3.12, LVGL 8.3.10 del ZIP del demo de Waveshare, ArduinoJson 7.4.3.

Estado: compila, y su parser y su diseño se revisaron en una computadora contra respuestas reales de la API. Todavía **no** se probó en la placa física; solo se probó el demo oficial de Waveshare.

Los drivers de pantalla están adaptados del demo oficial de Waveshare. Ver [`ai_quota_monitor/NOTICE.md`](ai_quota_monitor/NOTICE.md).
