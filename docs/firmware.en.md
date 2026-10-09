# Firmware and display — pending

**English** · [Español](firmware.md)

Target: `firmware/esp32-c6-lcd-1.47/`. There is no buildable firmware yet. Choose and verify a toolchain and libraries compatible with the ESP32-C6 and ST7789 before implementing; use the official Waveshare examples as a reference.

Planned configuration: SSID, Wi-Fi password, API URL, read token and time zone. Keep those values out of the repo. Provider credentials must not go on the ESP32.

Design:

- PERSONAL in light blue and WORK in violet.
- Automatic rotation every 10s; no fake touch buttons.
- One card per service, with remaining balance, separate quotas and reset time.
- OpenRouter in USD, without a monthly bar or an invented reset.
- Last update visible, distinct from the read time.
- Show MOCK when `source=mock`.
- Handle `null`, stale, auth_required, unavailable and missing Wi-Fi.
- Optional RGB LED for low balance; configurable thresholds.

Validate real legibility at 172 × 320: simplify or paginate if three cards turn out dense. Show the reset per window or state clearly which one it refers to. The firmware must validate TLS on HTTPS connections and keep the HTTP refresh separate from the screen refresh.
