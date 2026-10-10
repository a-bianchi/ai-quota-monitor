#include "net.h"
#include <Arduino.h>
#include <ArduinoJson.h>
#include <HTTPClient.h>
#include <NetworkClientSecure.h>
#include <WiFi.h>
#include <time.h>
#include "settings.h"
#include "usage_parser.h"

static const uint32_t POLL_INTERVAL_MS = 60 * 1000;
static const uint32_t RETRY_MIN_MS = 5 * 1000;
static const uint32_t RETRY_MAX_MS = 5 * 60 * 1000;
static const uint32_t WIFI_RETRY_MIN_MS = 15 * 1000;
static const uint32_t WIFI_RETRY_MAX_MS = 2 * 60 * 1000;
static const int32_t CONNECT_TIMEOUT_MS = 8000;
static const uint16_t READ_TIMEOUT_MS = 10000;
static const uint32_t CLOCK_TIMEOUT_MS = 20000;

static SemaphoreHandle_t lock;
static NetView shared;
static uint32_t version = 1;
// Parsed outside the lock, then copied in.
static Snapshot scratch;

static bool isHttps() {
  return strncmp(API_BASE_URL, "https://", 8) == 0;
}

static void publish(LinkState link, int httpCode, const Snapshot *data) {
  xSemaphoreTake(lock, portMAX_DELAY);
  if (data || shared.link != link || shared.httpCode != httpCode) {
    version++;
  }
  shared.link = link;
  shared.httpCode = httpCode;
  if (data) {
    // Earlier data is kept on errors; it is only replaced by a good read.
    shared.data = *data;
    shared.hasData = true;
    shared.lastOkMs = millis();
  }
  xSemaphoreGive(lock);
}

bool netRead(NetView &out, uint32_t &seenVersion) {
  bool changed = false;
  xSemaphoreTake(lock, portMAX_DELAY);
  if (version != seenVersion) {
    out = shared;
    seenVersion = version;
    changed = true;
  }
  xSemaphoreGive(lock);
  return changed;
}

static LinkState checkConfig() {
  if (strlen(WIFI_SSID) == 0 || strlen(DEVICE_TOKEN) == 0 || strlen(API_BASE_URL) == 0) {
    return LINK_CONFIG_MISSING;
  }
  const char *host = nullptr;
  if (strncmp(API_BASE_URL, "http://", 7) == 0) {
    host = API_BASE_URL + 7;
  } else if (isHttps()) {
    host = API_BASE_URL + 8;
  } else {
    return LINK_CONFIG_URL;
  }
  if (strncmp(host, "localhost", 9) == 0 || strncmp(host, "127.", 4) == 0 || strncmp(host, "[::1]", 5) == 0) {
    return LINK_CONFIG_LOCALHOST;
  }
  return LINK_OK;
}

static LinkState fetch(int &httpCode) {
  httpCode = 0;
  String url = API_BASE_URL;
  while (url.endsWith("/")) {
    url.remove(url.length() - 1);
  }
  url += "/usage";

  NetworkClient plain;
  NetworkClientSecure secure;
  HTTPClient http;
  http.setConnectTimeout(CONNECT_TIMEOUT_MS);
  http.setTimeout(READ_TIMEOUT_MS);
  http.setReuse(false);
  // HTTP/1.0 avoids chunked responses, so the body can be parsed as a stream.
  http.useHTTP10(true);

  bool started;
  if (isHttps()) {
    // The certificate is always validated: a private CA from config.h, or the
    // public authorities bundled with the core.
    if (strlen(API_CA_CERT) > 0) {
      secure.setCACert(API_CA_CERT);
    } else {
      secure.useBuiltinCACertBundle();
    }
    secure.setHandshakeTimeout(CONNECT_TIMEOUT_MS / 1000);
    started = http.begin(secure, url);
  } else {
    started = http.begin(plain, url);
  }
  if (!started) {
    return LINK_CONFIG_URL;
  }
  http.addHeader("Authorization", String("Bearer ") + DEVICE_TOKEN);
  http.addHeader("Accept", "application/json");
  http.addHeader("Accept-Language", uiEnglish() ? "en" : "es");
  if (strlen(API_TIMEZONE) > 0) {
    http.addHeader("X-Timezone", API_TIMEZONE);
  }

  httpCode = http.GET();
  LinkState result = LINK_OK;
  if (httpCode <= 0) {
    result = LINK_UNREACHABLE;
  } else if (httpCode == 401) {
    result = LINK_UNAUTHORIZED;
  } else if (httpCode != 200) {
    result = LINK_HTTP_ERROR;
  } else {
    JsonDocument filter;
    usageFilter(filter);
    JsonDocument doc;
    DeserializationError error = deserializeJson(doc, http.getStream(), DeserializationOption::Filter(filter),
                                                 DeserializationOption::NestingLimit(8));
    if (error || !parseUsage(doc.as<JsonVariantConst>(), scratch)) {
      Serial.printf("[api] invalid response: %s\n", error ? error.c_str() : "not a /usage document");
      result = LINK_BAD_RESPONSE;
    }
  }
  http.end();
  return result;
}

static bool clockReady() {
  return time(nullptr) > 1700000000;
}

static void netTask(void *) {
  LinkState config = checkConfig();
  if (config != LINK_OK) {
    Serial.println("[config] incomplete or invalid config.h");
    publish(config, 0, nullptr);
    vTaskDelete(nullptr);
    return;
  }

  WiFi.mode(WIFI_STA);
  WiFi.setAutoReconnect(true);
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
  publish(LINK_WIFI_DOWN, 0, nullptr);

  bool wasConnected = false;
  uint32_t wifiRetryMs = WIFI_RETRY_MIN_MS;
  uint32_t nextWifiAttempt = millis() + wifiRetryMs;
  uint32_t retryMs = RETRY_MIN_MS;
  uint32_t nextFetch = 0;
  uint32_t clockSince = 0;

  for (;;) {
    vTaskDelay(pdMS_TO_TICKS(200));
    uint32_t now = millis();

    if (WiFi.status() != WL_CONNECTED) {
      if (wasConnected) {
        Serial.println("[wifi] connection lost");
        wasConnected = false;
        wifiRetryMs = WIFI_RETRY_MIN_MS;
        nextWifiAttempt = now + wifiRetryMs;
      }
      publish(LINK_WIFI_DOWN, 0, nullptr);
      // The core reconnects by itself; this restarts the attempt if it stalls.
      if ((int32_t)(now - nextWifiAttempt) >= 0) {
        Serial.println("[wifi] retrying");
        WiFi.disconnect();
        WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
        wifiRetryMs = min(wifiRetryMs * 2, WIFI_RETRY_MAX_MS);
        nextWifiAttempt = now + wifiRetryMs;
      }
      continue;
    }
    if (!wasConnected) {
      wasConnected = true;
      Serial.printf("[wifi] connected, IP %s\n", WiFi.localIP().toString().c_str());
      retryMs = RETRY_MIN_MS;
      nextFetch = now;
      clockSince = now;
      if (isHttps()) {
        configTime(0, 0, "pool.ntp.org", "time.google.com");
      }
    }

    // Certificates cannot be validated without the current date.
    if (isHttps() && !clockReady()) {
      if (now - clockSince > CLOCK_TIMEOUT_MS) {
        publish(LINK_CLOCK, 0, nullptr);
      }
      continue;
    }

    if ((int32_t)(now - nextFetch) < 0) {
      continue;
    }
    int httpCode = 0;
    LinkState result = fetch(httpCode);
    if (result == LINK_OK) {
      publish(LINK_OK, httpCode, &scratch);
      retryMs = RETRY_MIN_MS;
      nextFetch = millis() + POLL_INTERVAL_MS;
      if (scratch.truncated) {
        Serial.println("[api] more accounts or quotas than the screen model holds; the rest is not shown");
      }
    } else {
      publish(result, httpCode, nullptr);
      Serial.printf("[api] read failed (state %d, HTTP %d); retrying in %lu s\n", (int)result, httpCode,
                    (unsigned long)(retryMs / 1000));
      nextFetch = millis() + retryMs;
      retryMs = min(retryMs * 2, RETRY_MAX_MS);
    }
  }
}

void netBegin() {
  lock = xSemaphoreCreateMutex();
  shared.link = LINK_STARTING;
  // TLS handshakes need more stack than the default task size.
  xTaskCreate(netTask, "net", 12288, nullptr, 1, nullptr);
}
