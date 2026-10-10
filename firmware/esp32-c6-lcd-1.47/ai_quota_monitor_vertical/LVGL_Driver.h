// LVGL 8.3 glue for the ST7789 display.
//
// Adapted from the official Waveshare demo (Arduino/examples/LVGL_Arduino in
// ESP32-C6-LCD-1.47-Demo.zip). See NOTICE.md. The board has no touch panel, so
// no input device is registered.
#pragma once

#include <lvgl.h>
#include "Display_ST7789.h"

#define LVGL_WIDTH LCD_WIDTH
#define LVGL_HEIGHT LCD_HEIGHT
#define LVGL_BUF_LEN (LVGL_WIDTH * LVGL_HEIGHT / 20)

#define EXAMPLE_LVGL_TICK_PERIOD_MS 5

void Lvgl_Init(void);
// Call often from loop(); it lets LVGL redraw.
void Timer_Loop(void);
