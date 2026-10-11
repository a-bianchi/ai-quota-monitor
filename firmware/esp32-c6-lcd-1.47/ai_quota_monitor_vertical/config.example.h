// Copy this file to config.h (same folder) and fill in your own values.
// config.h is ignored by Git: never commit it.
//
// Only the read token goes on the device. Do not put provider keys, Codex or
// Claude sessions, COLLECTOR_TOKEN or the panel password here.
#pragma once

// 2.4 GHz Wi-Fi network; the ESP32-C6 does not use 5 GHz networks.
#define WIFI_SSID ""
#define WIFI_PASSWORD ""

// Base URL of your own API, without a trailing slash. Use the LAN IP or the
// hostname of the machine that runs the server, with the port you published.
// "localhost" and "127.0.0.1" do not work: on the ESP32 they mean the ESP32.
#define API_BASE_URL "http://192.168.1.50:8080"

// DEVICE_TOKEN of that server (the value in its server/.env or platform variables).
#define DEVICE_TOKEN ""

// "es" or "en": interface texts and quota names.
#define UI_LANGUAGE "es"

// IANA time zone used for the times on screen.
#define API_TIMEZONE "America/Argentina/Buenos_Aires"

// Only for https:// URLs. Leave empty to trust the public certificate
// authorities bundled with the ESP32 core (Let's Encrypt and similar). For a
// private CA, paste its root certificate in PEM format. The certificate is
// always validated.
#define API_CA_CERT ""

// Backlight, 0 to 100. The manufacturer recommends 50 or lower for prolonged use.
#define DISPLAY_BRIGHTNESS 40

// RGB LED, 0 to 100; 0 turns it off. It shows the usage level of the page on
// screen: green for low, orange for medium, red for high.
#define LED_BRIGHTNESS 10

// If the LED shows red where green is expected, use LED_COLOR_ORDER_GRB.
#define LED_COLOR_ORDER LED_COLOR_ORDER_RGB
