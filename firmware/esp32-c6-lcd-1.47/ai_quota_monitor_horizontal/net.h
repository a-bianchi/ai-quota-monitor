// Wi-Fi and API polling. Everything here runs in its own task so the screen
// keeps refreshing while a request is in progress.
#pragma once
#include <stdint.h>
#include "model.h"

enum LinkState : uint8_t {
  LINK_STARTING,
  LINK_CONFIG_MISSING,    // config.h still has empty values
  LINK_CONFIG_URL,        // API_BASE_URL is not http:// or https://
  LINK_CONFIG_LOCALHOST,  // API_BASE_URL points to the ESP32 itself
  LINK_WIFI_DOWN,
  LINK_CLOCK,         // waiting for the time needed to validate certificates
  LINK_OK,
  LINK_UNREACHABLE,   // no TCP/TLS connection or no answer in time
  LINK_UNAUTHORIZED,  // HTTP 401
  LINK_HTTP_ERROR,    // any other HTTP status
  LINK_BAD_RESPONSE,  // the body is not a /usage response
};

struct NetView {
  LinkState link;
  int httpCode;
  bool hasData;
  // millis() of the last successful read.
  uint32_t lastOkMs;
  Snapshot data;
};

void netBegin();

// Copies the current state into out when it changed since seenVersion.
bool netRead(NetView &out, uint32_t &seenVersion);
