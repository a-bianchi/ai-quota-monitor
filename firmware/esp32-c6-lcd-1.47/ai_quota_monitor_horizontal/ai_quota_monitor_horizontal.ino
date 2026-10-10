// AI Quota Monitor for the Waveshare ESP32-C6-LCD-1.47 (no touch).
//
// Reads GET /usage from your own server and shows the personal and work quotas.
// Horizontal sketch: 320 x 172 screen. layout.h holds what differs from
// ai_quota_monitor_vertical; every other file is the same in both.
// Setup, libraries and flashing: docs/firmware.md in the repository.
//
// Board: ESP32C6 Dev Module. Partition Scheme: Huge APP (3MB No OTA/1MB SPIFFS).
#include "Display_ST7789.h"
#include "LVGL_Driver.h"
#include "net.h"
#include "settings.h"
#include "ui.h"

void setup() {
  Serial.begin(115200);
  LCD_Init();
  Set_Backlight(DISPLAY_BRIGHTNESS);
  Lvgl_Init();
  uiBegin();
  netBegin();
}

void loop() {
  // Network requests run in their own task (net.cpp); this loop only draws.
  uiLoop();
  Timer_Loop();
  delay(5);
}
