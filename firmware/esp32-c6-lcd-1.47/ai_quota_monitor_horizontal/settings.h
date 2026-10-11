// Loads the local configuration and fails with a clear message without it.
#pragma once
#include <string.h>

#if __has_include("config.h")
#include "config.h"
#else
#error "Missing config.h: copy config.example.h to config.h in this folder and fill it in. See docs/firmware.md"
#endif

#ifndef UI_LANGUAGE
#define UI_LANGUAGE "es"
#endif
#ifndef API_TIMEZONE
#define API_TIMEZONE ""
#endif
#ifndef API_CA_CERT
#define API_CA_CERT ""
#endif
#ifndef DISPLAY_BRIGHTNESS
#define DISPLAY_BRIGHTNESS 40
#endif
#ifndef LED_BRIGHTNESS
#define LED_BRIGHTNESS 10
#endif
#ifndef LED_COLOR_ORDER
#define LED_COLOR_ORDER LED_COLOR_ORDER_RGB
#endif

static inline bool uiEnglish() {
  return strncmp(UI_LANGUAGE, "en", 2) == 0;
}
