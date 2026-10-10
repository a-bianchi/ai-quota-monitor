// Data shown on screen, as parsed from GET /usage.
#pragma once
#include <stdint.h>

#define MAX_ACCOUNTS 16
#define MAX_QUOTAS 48
#define NAME_LEN 28
#define TIME_LEN 12  // "dd/mm HH:MM"

enum Level : uint8_t { LEVEL_UNKNOWN, LEVEL_LOW, LEVEL_MEDIUM, LEVEL_HIGH };
enum Status : uint8_t { STATUS_OK, STATUS_STALE, STATUS_AUTH_REQUIRED, STATUS_UNAVAILABLE };
enum Unit : uint8_t { UNIT_PERCENT, UNIT_CREDITS, UNIT_USD, UNIT_TOKENS, UNIT_OTHER };

struct Quota {
  char name[NAME_LEN];
  Unit unit;
  Level level;
  bool hasUsed;
  bool hasLimit;
  double used;
  double limit;
  char reset[TIME_LEN];  // empty when the API sends null
};

struct Account {
  char name[NAME_LEN];
  bool work;
  bool mock;
  Status status;
  char updated[TIME_LEN];  // empty when the API sends null
  uint8_t firstQuota;
  uint8_t quotaCount;
};

struct Snapshot {
  uint8_t accountCount;
  uint8_t quotaCount;
  // Set when the response had more accounts or quotas than fit above.
  bool truncated;
  char readAt[6];  // "HH:MM" of generatedAt
  Account accounts[MAX_ACCOUNTS];
  Quota quotas[MAX_QUOTAS];
};
