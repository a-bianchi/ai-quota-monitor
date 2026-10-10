#include "usage_parser.h"
#include <string.h>

void usageFilter(JsonDocument &filter) {
  filter["generatedAt"] = true;
  JsonObject account = filter["accounts"][0].to<JsonObject>();
  account["name"] = true;
  account["scope"] = true;
  account["source"] = true;
  account["status"] = true;
  account["updatedAt"] = true;
  JsonObject quota = account["quotas"][0].to<JsonObject>();
  quota["name"] = true;
  quota["unit"] = true;
  quota["used"] = true;
  quota["limit"] = true;
  quota["resetsAt"] = true;
  quota["usageLevel"] = true;
}

void foldText(const char *in, char *out, size_t size) {
  // Latin-1 letters U+00C0..U+00FF without their accents.
  static const char latin[] =
    "AAAAAAACEEEEIIII"
    "DNOOOOOxOUUUUYTs"
    "aaaaaaaceeeeiiii"
    "dnooooo/ouuuuyty";
  size_t n = 0;
  if (size == 0) {
    return;
  }
  auto put = [&](const char *s) {
    while (*s && n + 1 < size) {
      out[n++] = *s++;
    }
  };
  const unsigned char *p = (const unsigned char *)(in ? in : "");
  while (*p && n + 1 < size) {
    if (*p < 0x80) {
      out[n++] = (*p < 0x20) ? ' ' : (char)*p;
      p++;
      continue;
    }
    unsigned code = '?';
    int extra = 0;
    if ((*p & 0xE0) == 0xC0) {
      code = *p & 0x1F;
      extra = 1;
    } else if ((*p & 0xF0) == 0xE0) {
      code = *p & 0x0F;
      extra = 2;
    } else if ((*p & 0xF8) == 0xF0) {
      code = *p & 0x07;
      extra = 3;
    }
    p++;
    while (extra-- > 0 && (*p & 0xC0) == 0x80) {
      code = (code << 6) | (*p & 0x3F);
      p++;
    }
    if (code >= 0xC0 && code <= 0xFF) {
      char c[2] = { latin[code - 0xC0], 0 };
      put(c);
    } else if (code == 0xB7 || code == 0x2022 || code == 0x2013 || code == 0x2014) {
      put("-");  // middle dot, bullet and dashes
    } else if (code == 0x2026) {
      put("...");
    } else if (code == 0xA0) {
      put(" ");
    } else if (code == 0xBF || code == 0xA1) {
      // Inverted punctuation is dropped.
    } else {
      put("?");
    }
  }
  out[n] = 0;
}

static bool digits(const char *s, int from, int count) {
  for (int i = from; i < from + count; i++) {
    if (s[i] < '0' || s[i] > '9') {
      return false;
    }
  }
  return true;
}

// The API already converts the times to the requested zone, so the local date
// and time are read straight from the RFC 3339 text: "2026-10-08T17:40:00-03:00".
static bool validTime(const char *iso) {
  return iso && strlen(iso) >= 16 && digits(iso, 0, 4) && iso[4] == '-' && digits(iso, 5, 2) && iso[7] == '-'
         && digits(iso, 8, 2) && iso[10] == 'T' && digits(iso, 11, 2) && iso[13] == ':' && digits(iso, 14, 2);
}

// Writes "HH:MM" for a time on the same day as today, otherwise "dd/mm HH:MM".
static void shortTime(const char *iso, const char *today, char out[TIME_LEN]) {
  out[0] = 0;
  if (!validTime(iso)) {
    return;
  }
  size_t n = 0;
  if (!validTime(today) || strncmp(iso, today, 10) != 0) {
    out[n++] = iso[8];
    out[n++] = iso[9];
    out[n++] = '/';
    out[n++] = iso[5];
    out[n++] = iso[6];
    out[n++] = ' ';
  }
  memcpy(out + n, iso + 11, 5);
  out[n + 5] = 0;
}

static bool is(const char *value, const char *expected) {
  return value && strcmp(value, expected) == 0;
}

bool parseUsage(JsonVariantConst root, Snapshot &out) {
  memset(&out, 0, sizeof(out));
  JsonArrayConst accounts = root["accounts"].as<JsonArrayConst>();
  if (accounts.isNull()) {
    return false;
  }
  const char *today = root["generatedAt"].as<const char *>();
  if (validTime(today)) {
    memcpy(out.readAt, today + 11, 5);
  }
  for (JsonVariantConst item : accounts) {
    const char *scope = item["scope"].as<const char *>();
    if (!is(scope, "personal") && !is(scope, "work")) {
      continue;
    }
    if (out.accountCount >= MAX_ACCOUNTS) {
      out.truncated = true;
      break;
    }
    Account &a = out.accounts[out.accountCount++];
    foldText(item["name"].as<const char *>(), a.name, sizeof(a.name));
    a.work = is(scope, "work");
    a.mock = is(item["source"].as<const char *>(), "mock");
    const char *status = item["status"].as<const char *>();
    // An unknown status is treated as unavailable, like the web panel does.
    a.status = is(status, "ok") ? STATUS_OK
               : is(status, "stale") ? STATUS_STALE
               : is(status, "auth_required") ? STATUS_AUTH_REQUIRED
               : STATUS_UNAVAILABLE;
    shortTime(item["updatedAt"].as<const char *>(), today, a.updated);
    a.firstQuota = out.quotaCount;
    for (JsonVariantConst q : item["quotas"].as<JsonArrayConst>()) {
      if (out.quotaCount >= MAX_QUOTAS) {
        out.truncated = true;
        break;
      }
      Quota &quota = out.quotas[out.quotaCount++];
      a.quotaCount++;
      foldText(q["name"].as<const char *>(), quota.name, sizeof(quota.name));
      const char *unit = q["unit"].as<const char *>();
      quota.unit = is(unit, "percent") ? UNIT_PERCENT
                   : is(unit, "usd") ? UNIT_USD
                   : is(unit, "credits") ? UNIT_CREDITS
                   : is(unit, "tokens") ? UNIT_TOKENS
                   : UNIT_OTHER;
      // null and missing values stay missing; no number is invented.
      if (q["used"].is<double>()) {
        quota.hasUsed = true;
        quota.used = q["used"].as<double>();
      }
      if (q["limit"].is<double>()) {
        quota.hasLimit = true;
        quota.limit = q["limit"].as<double>();
      }
      const char *level = q["usageLevel"].as<const char *>();
      quota.level = is(level, "low") ? LEVEL_LOW
                    : is(level, "medium") ? LEVEL_MEDIUM
                    : is(level, "high") ? LEVEL_HIGH
                    : LEVEL_UNKNOWN;
      shortTime(q["resetsAt"].as<const char *>(), today, quota.reset);
    }
  }
  return true;
}
