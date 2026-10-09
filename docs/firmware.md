# Firmware and display — pending

**English** · [Español](firmware.es.md)

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

## What exists today

- The API the firmware will consume: `GET /usage` with `DEVICE_TOKEN`. See the [reference](api.md).
- Each quota's `usageLevel` field, meant for choosing the color without computing percentages on the device.
- The **Display** view of the [web panel](panel.md), which previews the planned content at 172 × 320 px with the API data.

## Next steps

1. Choose the toolchain and the graphics library, and verify they support ESP32-C6 and ST7789.
2. Light up the screen on the real board with an official Waveshare example.
3. Connect to Wi-Fi, read `GET /usage` and draw one page.
4. Validate TLS, error handling and legibility before publishing flashing instructions.
