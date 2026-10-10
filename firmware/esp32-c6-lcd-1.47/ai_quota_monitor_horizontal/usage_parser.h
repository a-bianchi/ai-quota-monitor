// Parser for the GET /usage contract (schemaVersion 1). It depends only on
// ArduinoJson, so it can also be compiled and tested on a computer.
#pragma once
#include <ArduinoJson.h>
#include <stddef.h>
#include "model.h"

// Fills the filter that keeps only the fields the screen uses.
void usageFilter(JsonDocument &filter);

// Returns false when the document is not a /usage response.
bool parseUsage(JsonVariantConst root, Snapshot &out);

// Copies UTF-8 text as ASCII: the built-in LVGL fonts have no accented letters.
void foldText(const char *in, char *out, size_t size);
