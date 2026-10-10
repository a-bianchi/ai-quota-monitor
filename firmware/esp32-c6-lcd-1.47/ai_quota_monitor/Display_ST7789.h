// ST7789 driver for the Waveshare ESP32-C6-LCD-1.47 (172 x 320, portrait).
//
// Adapted from the official Waveshare demo (Arduino/examples/LVGL_Arduino in
// ESP32-C6-LCD-1.47-Demo.zip). Pins, SPI settings and the panel initialisation
// sequence are the manufacturer's. See NOTICE.md.
#pragma once
#include <Arduino.h>
#include <SPI.h>

#define LCD_WIDTH 172
#define LCD_HEIGHT 320

#define SPIFreq 80000000
#define EXAMPLE_PIN_NUM_MISO 5
#define EXAMPLE_PIN_NUM_MOSI 6
#define EXAMPLE_PIN_NUM_SCLK 7
#define EXAMPLE_PIN_NUM_LCD_CS 14
#define EXAMPLE_PIN_NUM_LCD_DC 15
#define EXAMPLE_PIN_NUM_LCD_RST 21
#define EXAMPLE_PIN_NUM_BK_LIGHT 22
#define Frequency 1000
#define Resolution 10

// The 172 px wide glass is centred in the controller's 240 px memory.
#define Offset_X 34
#define Offset_Y 0

void LCD_Init(void);
void LCD_SetCursor(uint16_t Xstart, uint16_t Ystart, uint16_t Xend, uint16_t Yend);
void LCD_addWindow(uint16_t Xstart, uint16_t Ystart, uint16_t Xend, uint16_t Yend, uint16_t *color);

void Backlight_Init(void);
// Light is a percentage from 0 to 100.
void Set_Backlight(uint8_t Light);
