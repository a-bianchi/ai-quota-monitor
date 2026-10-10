#include "ui.h"
#include <Arduino.h>
#include <lvgl.h>
#include <stdio.h>
#include "Display_ST7789.h"
#include "logos.h"
#include "net.h"
#include "settings.h"

// Layout for 172 x 320 or 320 x 172 (layout.h), following the Display view of
// the web panel.
static const int SCREEN_W = LCD_WIDTH;
static const int SCREEN_H = LCD_HEIGHT;

#if UI_ONE_ACCOUNT
// One account per page: 16 px titles, 14 px quotas, 12 px details. The account
// shows a large logo with its name and provider beside it.
#define FONT_TEXT (&lv_font_montserrat_14)
#define FONT_SMALL (&lv_font_montserrat_12)
#if LV_FONT_MONTSERRAT_16
#define FONT_TITLE (&lv_font_montserrat_16)
static const int TITLE_H = 18;  // line height of the 16 px font
#else  // 16 px is not enabled in lv_conf.h
#define FONT_TITLE FONT_TEXT
static const int TITLE_H = 16;
#endif
#define FONT_NAME FONT_TITLE
static const int NAME_H = TITLE_H;
static const int ROW_H = 16;    // line height of the 14 px font
static const int SMALL_H = 15;  // line height of the 12 px font
static const int PAD_X = 8;
static const int BAR_H = 6;
static const int CARD_GAP = 0;  // nothing follows the only account
static const int LOGO = LOGO_LARGE;
static const int LOGO_GAP = 8;
static const int ACCOUNT_H = 32;  // logo, or name and provider lines
#else
// Every account of a scope: 14 px title, 12 px for the rest.
#define FONT_TEXT (&lv_font_montserrat_12)
#define FONT_SMALL FONT_TEXT
#define FONT_TITLE (&lv_font_montserrat_14)
#define FONT_NAME FONT_TEXT
static const int TITLE_H = 16;  // line height of the 14 px font
static const int NAME_H = 15;
static const int ROW_H = 15;    // line height of the 12 px font
static const int SMALL_H = 15;
static const int PAD_X = 6;
static const int BAR_H = 4;
static const int CARD_GAP = 9;  // space, separator line and space
static const int LOGO = LOGO_SMALL;
static const int LOGO_GAP = 4;
static const int ACCOUNT_H = NAME_H;
#endif

static const int PAD_Y = DISPLAY_LANDSCAPE ? 6 : 8;
static const int INNER_W = SCREEN_W - 2 * PAD_X;
static const int COL_GAP = 12;
static const int CONTENT_TOP = PAD_Y + TITLE_H + 7;
static const int FOOTER_Y = SCREEN_H - PAD_Y - SMALL_H;
static const int CONTENT_H = FOOTER_Y - 3 - CONTENT_TOP;
// Landscape is short: with one account per page, its update line goes in the footer.
static const bool TAIL_IN_FOOTER = DISPLAY_LANDSCAPE && UI_ONE_ACCOUNT;

static const uint32_t ROTATE_MS = 10 * 1000;
// Data kept after failed reads is marked as stale once it is this old.
static const uint32_t STALE_MARK_MS = 150 * 1000;

// Same palette as the panel; usageLevel picks green, orange, red or gray.
#define COLOR_TEXT lv_color_hex(0xFFFFFF)
#define COLOR_MUTED lv_color_hex(0x9AA7B4)
#define COLOR_LINE lv_color_hex(0x2B3440)
#define COLOR_PERSONAL lv_color_hex(0x6CC4F5)
#define COLOR_WORK lv_color_hex(0xB48CF2)
#define COLOR_ACCENT lv_color_hex(0x58A6FF)
#define COLOR_LOW lv_color_hex(0x3FB950)
#define COLOR_MEDIUM lv_color_hex(0xF0883E)
#define COLOR_HIGH lv_color_hex(0xF85149)
#define COLOR_UNKNOWN lv_color_hex(0x8B949E)

struct Chunk {
  uint8_t account;
  uint8_t firstQuota;
  uint8_t quotaCount;
  bool tail;  // last part of the account: shows the update line
  uint16_t height;
};
struct Page {
  bool work;
  uint8_t firstChunk;
  uint8_t chunkCount;
};

static const int MAX_CHUNKS = MAX_QUOTAS + MAX_ACCOUNTS;
static const int MAX_PAGES = MAX_CHUNKS + 2;

static NetView view;
static uint32_t seenVersion = 0;
static Chunk chunks[MAX_CHUNKS];
static Page pages[MAX_PAGES];
static int chunkCount = 0;
static int pageCount = 0;
static int pageIndex = 0;
static uint32_t lastRotate = 0;
static bool markedStale = false;

static const char *tr(const char *es, const char *en) {
  return uiEnglish() ? en : es;
}

static lv_color_t levelColor(Level level) {
  switch (level) {
    case LEVEL_LOW: return COLOR_LOW;
    case LEVEL_MEDIUM: return COLOR_MEDIUM;
    case LEVEL_HIGH: return COLOR_HIGH;
    default: return COLOR_UNKNOWN;
  }
}

static int textWidth(const char *text, const lv_font_t *font) {
  return lv_txt_get_width(text, strlen(text), font, 0, LV_TEXT_FLAG_NONE);
}

// One line of text clipped with dots to the given width.
static lv_obj_t *line(const char *text, const lv_font_t *font, lv_color_t color, int x, int y, int w,
                      lv_text_align_t align = LV_TEXT_ALIGN_LEFT) {
  lv_obj_t *label = lv_label_create(lv_scr_act());
  lv_obj_set_style_text_font(label, font, 0);
  lv_obj_set_style_text_color(label, color, 0);
  lv_obj_set_style_text_align(label, align, 0);
  lv_label_set_long_mode(label, LV_LABEL_LONG_DOT);
  lv_obj_set_size(label, w, lv_font_get_line_height(font));
  lv_obj_set_pos(label, x, y);
  lv_label_set_text(label, text);
  return label;
}

// Text on the left and a value on the right, inside w pixels from x; the value
// is never clipped. A smaller right font sits on the baseline of the left one.
static void pair(const char *left, const lv_font_t *leftFont, lv_color_t leftColor, const char *right,
                 const lv_font_t *rightFont, lv_color_t rightColor, int x, int y, int w) {
  int rightW = min(textWidth(right, rightFont) + 1, w);
  if (rightW > 1) {
    int drop = max(lv_font_get_line_height(leftFont) - lv_font_get_line_height(rightFont), 0);
    line(right, rightFont, rightColor, x + w - rightW, y + drop, rightW, LV_TEXT_ALIGN_RIGHT);
  }
  int leftW = w - rightW - 4;
  if (leftW > 8) {
    line(left, leftFont, leftColor, x, y, leftW);
  }
}

static void rect(int x, int y, int w, int h, lv_color_t color, int radius = 0) {
  lv_obj_t *obj = lv_obj_create(lv_scr_act());
  lv_obj_remove_style_all(obj);
  lv_obj_set_style_bg_color(obj, color, 0);
  lv_obj_set_style_bg_opa(obj, LV_OPA_COVER, 0);
  lv_obj_set_style_radius(obj, radius, 0);
  lv_obj_set_pos(obj, x, y);
  lv_obj_set_size(obj, w, h);
}

static void number(char *out, size_t size, double value) {
  if (value >= 1e6) {
    snprintf(out, size, "%.1fM", value / 1e6);
  } else if (value >= 1e4) {
    snprintf(out, size, "%.1fk", value / 1e3);
  } else if (value == (long)value) {
    snprintf(out, size, "%ld", (long)value);
  } else {
    snprintf(out, size, "%.1f", value);
  }
}

static void amount(char *out, size_t size, Unit unit, double value) {
  char n[16];
  number(n, sizeof(n), value);
  switch (unit) {
    case UNIT_USD: snprintf(out, size, "$%.2f", value); break;
    case UNIT_PERCENT: snprintf(out, size, "%s%%", n); break;
    case UNIT_CREDITS: snprintf(out, size, "%s cred.", n); break;
    case UNIT_TOKENS: snprintf(out, size, "%s tok.", n); break;
    default: snprintf(out, size, "%s", n); break;
  }
}

// What is left of a quota, as in the panel. Missing values are shown as
// missing: nothing is computed without both numbers.
static void quotaValue(char *out, size_t size, const Quota &q) {
  char a[20];
  if (q.hasUsed && q.hasLimit) {
    if (q.used >= q.limit) {
      snprintf(out, size, "%s", tr("Agotado", "Exhausted"));
    } else {
      amount(out, size, q.unit, q.limit - q.used);
    }
  } else if (q.hasUsed) {
    // No limit, for example OpenRouter without a known credit total.
    amount(a, sizeof(a), q.unit, q.used);
    snprintf(out, size, "%s %s", a, tr("usado", "used"));
  } else {
    snprintf(out, size, "--");
  }
}

static bool hasBar(const Quota &q) {
  return q.hasUsed && q.hasLimit && q.limit > 0;
}

static int quotaHeight(const Quota &q) {
  return 2 + ROW_H + (hasBar(q) ? BAR_H + 3 : 0) + (q.reset[0] ? SMALL_H : 0);
}

// Landscape has room for two quotas side by side.
static int columns(const Account &a) {
  return DISPLAY_LANDSCAPE && a.quotaCount > 2 ? 2 : 1;
}

static bool hasTail(const Account &a) {
  return a.updated[0] || a.mock;
}

// Splits the accounts of one scope into pages. An account that is taller than
// the screen continues on the next page. With UI_ONE_ACCOUNT every account
// starts its own page.
static void paginate(bool work) {
  int firstPage = pageCount;
  int used = CONTENT_H + 1;  // forces the first page
  for (int i = 0; i < view.data.accountCount; i++) {
    const Account &a = view.data.accounts[i];
    if (a.work != work) {
      continue;
    }
    int cols = columns(a);
    int next = 0;
    do {
      if (chunkCount >= MAX_CHUNKS) {
        return;
      }
      Chunk c = { (uint8_t)i, (uint8_t)(a.firstQuota + next), 0, false, (uint16_t)(ACCOUNT_H + CARD_GAP) };
      if (a.quotaCount == 0) {
        c.height += ROW_H;
      }
      while (next < a.quotaCount) {
        // One row: as tall as its tallest quota.
        int count = min(cols, a.quotaCount - next);
        int h = 0;
        for (int k = 0; k < count; k++) {
          h = max(h, quotaHeight(view.data.quotas[a.firstQuota + next + k]));
        }
        if (c.quotaCount > 0 && c.height + h > CONTENT_H) {
          break;
        }
        c.height += h;
        c.quotaCount += count;
        next += count;
      }
      if (next >= a.quotaCount) {
        c.tail = true;
        if (hasTail(a) && !TAIL_IN_FOOTER) {
          c.height += 2 + SMALL_H;
        }
      }
      if (UI_ONE_ACCOUNT || used + c.height > CONTENT_H) {
        if (pageCount >= MAX_PAGES) {
          return;
        }
        pages[pageCount++] = { work, (uint8_t)chunkCount, 0 };
        used = 0;
      }
      chunks[chunkCount++] = c;
      pages[pageCount - 1].chunkCount++;
      used += c.height;
    } while (next < a.quotaCount);
  }
  (void)firstPage;
}

static void buildPages() {
  chunkCount = 0;
  pageCount = 0;
  if (!view.hasData) {
    return;
  }
  paginate(false);
  paginate(true);
  if (pageCount == 0) {
    // No accounts at all: one empty page per scope, like the panel.
    pages[pageCount++] = { false, 0, 0 };
    pages[pageCount++] = { true, 0, 0 };
  }
}

static const char *linkText(char *buffer, size_t size) {
  switch (view.link) {
    case LINK_STARTING: return tr("Iniciando...", "Starting...");
    case LINK_CONFIG_MISSING: return tr("Falta configurar", "Not configured");
    case LINK_CONFIG_URL: return tr("URL invalida", "Invalid URL");
    case LINK_CONFIG_LOCALHOST: return tr("localhost no sirve", "localhost will not work");
    case LINK_WIFI_DOWN: return tr("Sin Wi-Fi", "No Wi-Fi");
    case LINK_CLOCK: return tr("Sin hora (NTP)", "No time (NTP)");
    case LINK_UNREACHABLE: return tr("Sin conexion a la API", "API unreachable");
    case LINK_UNAUTHORIZED: return tr("Token rechazado (401)", "Token rejected (401)");
    case LINK_HTTP_ERROR:
      snprintf(buffer, size, tr("Error HTTP %d", "HTTP error %d"), view.httpCode);
      return buffer;
    case LINK_BAD_RESPONSE: return tr("Respuesta invalida", "Invalid response");
    default: return "";
  }
}

static const char *linkHint() {
  switch (view.link) {
    case LINK_CONFIG_MISSING:
      return tr("Completar WIFI_SSID, API_BASE_URL y DEVICE_TOKEN en config.h y volver a subir.",
                "Fill in WIFI_SSID, API_BASE_URL and DEVICE_TOKEN in config.h and upload again.");
    case LINK_CONFIG_URL:
      return tr("API_BASE_URL debe empezar con http:// o https://",
                "API_BASE_URL must start with http:// or https://");
    case LINK_CONFIG_LOCALHOST:
      return tr("En la ESP32, localhost es la propia placa. Usar la IP LAN o el hostname del servidor en API_BASE_URL.",
                "On the ESP32, localhost is the board itself. Use the server's LAN IP or hostname in API_BASE_URL.");
    case LINK_WIFI_DOWN:
      return tr("Conectando. Revisar nombre y clave de la red, y que sea de 2.4 GHz.",
                "Connecting. Check the network name and password, and that it is 2.4 GHz.");
    case LINK_CLOCK:
      return tr("Se necesita la hora por NTP para validar el certificado HTTPS.",
                "The time from NTP is needed to validate the HTTPS certificate.");
    case LINK_UNREACHABLE:
      return tr("Revisar URL y puerto, que el servidor este encendido y que acepte conexiones desde la red.",
                "Check the URL and port, that the server is running and that it accepts connections from the network.");
    case LINK_UNAUTHORIZED:
      return tr("Revisar DEVICE_TOKEN en config.h: debe ser el del servidor.",
                "Check DEVICE_TOKEN in config.h: it must be the server's.");
    case LINK_HTTP_ERROR:
      return tr("El servidor respondio con un error. Revisar sus logs.", "The server answered with an error. Check its logs.");
    case LINK_BAD_RESPONSE:
      return tr("La URL no parece ser la API del monitor.", "The URL does not look like the monitor's API.");
    default:
      return "";
  }
}

static lv_color_t linkColor() {
  switch (view.link) {
    case LINK_STARTING:
    case LINK_WIFI_DOWN:
    case LINK_CLOCK:
    case LINK_UNREACHABLE: return COLOR_MEDIUM;
    default: return COLOR_HIGH;
  }
}

// Full-screen message while there is nothing to show yet.
static void renderMessage() {
  char buffer[32];
  line("AI Quota Monitor", FONT_TITLE, COLOR_PERSONAL, PAD_X, PAD_Y, INNER_W);
  rect(PAD_X, PAD_Y + TITLE_H + 2, INNER_W, 1, COLOR_PERSONAL);
  int y = DISPLAY_LANDSCAPE ? 44 : 96;
  line(linkText(buffer, sizeof(buffer)), FONT_TITLE, linkColor(), PAD_X, y, INNER_W, LV_TEXT_ALIGN_CENTER);
  lv_obj_t *hint = lv_label_create(lv_scr_act());
  lv_obj_set_style_text_font(hint, FONT_SMALL, 0);
  lv_obj_set_style_text_color(hint, COLOR_MUTED, 0);
  lv_obj_set_style_text_align(hint, LV_TEXT_ALIGN_CENTER, 0);
  lv_label_set_long_mode(hint, LV_LABEL_LONG_WRAP);
  lv_obj_set_width(hint, INNER_W);
  lv_obj_set_pos(hint, PAD_X, y + 28);
  lv_label_set_text(hint, linkHint());
}

static const char *providerName(Provider provider) {
  switch (provider) {
    case PROVIDER_ANTHROPIC: return "Claude";
    case PROVIDER_OPENAI: return "Codex (ChatGPT)";
    case PROVIDER_CURSOR: return "Cursor";
    case PROVIDER_OPENROUTER: return "OpenRouter";
    case PROVIDER_KIRO: return "Kiro";
    default: return "";
  }
}

// Provider logo, or a tile with the initial of the account when it has none.
static void renderLogo(const Account &a, int x, int y) {
  const lv_img_dsc_t *image = logoFor(a.provider, UI_ONE_ACCOUNT);
  if (image) {
    lv_obj_t *img = lv_img_create(lv_scr_act());
    lv_img_set_src(img, image);
    lv_obj_set_pos(img, x, y);
    return;
  }
  const lv_font_t *font = UI_ONE_ACCOUNT ? FONT_TITLE : FONT_SMALL;
  char initial[2] = { a.name[0] ? a.name[0] : '?', 0 };
  if (initial[0] >= 'a' && initial[0] <= 'z') {
    initial[0] -= 'a' - 'A';
  }
  rect(x, y, LOGO, LOGO, COLOR_LINE, LOGO / 4);
  line(initial, font, COLOR_TEXT, x, y + (LOGO - lv_font_get_line_height(font)) / 2, LOGO, LV_TEXT_ALIGN_CENTER);
}

// Logo, account name and status; with UI_ONE_ACCOUNT, the provider below the name.
static void renderStatus(const Account &a, int y) {
  Status status = a.status;
  if (markedStale && status == STATUS_OK) {
    status = STATUS_STALE;
  }
  char text[32] = "";
  lv_color_t color = COLOR_TEXT;
  switch (status) {
    case STATUS_STALE:
      snprintf(text, sizeof(text), LV_SYMBOL_REFRESH " %s", tr("Desact.", "Stale"));
      color = COLOR_MEDIUM;
      break;
    case STATUS_AUTH_REQUIRED:
      snprintf(text, sizeof(text), LV_SYMBOL_WARNING " %s", tr("Sin sesion", "Login"));
      color = COLOR_ACCENT;
      break;
    case STATUS_UNAVAILABLE:
      snprintf(text, sizeof(text), LV_SYMBOL_CLOSE " %s", tr("Sin datos", "No data"));
      color = COLOR_HIGH;
      break;
    default:
      break;
  }
  int x = PAD_X + LOGO + LOGO_GAP;
  int w = INNER_W - LOGO - LOGO_GAP;
  renderLogo(a, PAD_X, y + (ACCOUNT_H - LOGO) / 2);
  pair(a.name, FONT_NAME, COLOR_TEXT, text, FONT_TEXT, color, x, y, w);
  if (UI_ONE_ACCOUNT) {
    line(providerName(a.provider), FONT_SMALL, COLOR_MUTED, x, y + ACCOUNT_H - SMALL_H, w);
  }
}

static void renderQuota(const Quota &q, int x, int y, int w) {
  char value[40];
  quotaValue(value, sizeof(value), q);
  y += 2;
  pair(q.name, FONT_TEXT, COLOR_MUTED, value, FONT_TEXT, levelColor(q.level), x, y, w);
  y += ROW_H;
  if (hasBar(q)) {
    double percent = q.used / q.limit * 100;
    percent = percent < 0 ? 0 : percent > 100 ? 100 : percent;
    int filled = (int)(w * percent / 100 + 0.5);
    rect(x, y + 1, w, BAR_H, COLOR_LINE, 2);
    if (filled > 0) {
      rect(x, y + 1, max(filled, 2), BAR_H, levelColor(q.level), 2);
    }
    y += BAR_H + 3;
  }
  if (q.reset[0]) {
    char reset[32];
    snprintf(reset, sizeof(reset), LV_SYMBOL_REFRESH " %s", q.reset);
    line(reset, FONT_SMALL, COLOR_MUTED, x, y, w);
  }
}

// "Act. 10/10 14:30", empty when the account was never updated.
static void updatedText(char *out, size_t size, const Account &a) {
  out[0] = 0;
  if (a.updated[0]) {
    snprintf(out, size, tr("Act. %s", "Upd. %s"), a.updated);
  }
}

static int renderChunk(const Chunk &c, int y) {
  const Account &a = view.data.accounts[c.account];
  renderStatus(a, y);
  y += ACCOUNT_H;
  if (a.quotaCount == 0) {
    line(tr("Sin cupos todavia", "No quotas yet"), FONT_TEXT, COLOR_MUTED, PAD_X, y, INNER_W);
    y += ROW_H;
  }
  int cols = columns(a);
  int colW = cols > 1 ? (INNER_W - COL_GAP * (cols - 1)) / cols : INNER_W;
  for (int i = 0; i < c.quotaCount; i += cols) {
    int h = 0;
    for (int k = 0; k < cols && i + k < c.quotaCount; k++) {
      const Quota &q = view.data.quotas[c.firstQuota + i + k];
      renderQuota(q, PAD_X + k * (colW + COL_GAP), y, colW);
      h = max(h, quotaHeight(q));
    }
    y += h;
  }
  if (c.tail && hasTail(a) && !TAIL_IN_FOOTER) {
    char updated[32];
    updatedText(updated, sizeof(updated), a);
    y += 2;
    pair(updated, FONT_SMALL, COLOR_MUTED, a.mock ? "MOCK" : "", FONT_SMALL, COLOR_MEDIUM, PAD_X, y, INNER_W);
    y += SMALL_H;
  }
  if (CARD_GAP > 0) {
    rect(PAD_X, y + 4, INNER_W, 1, lv_color_hex(0x222222));
  }
  return y + CARD_GAP;
}

static void renderPage() {
  const Page &page = pages[pageIndex];
  lv_color_t scope = page.work ? COLOR_WORK : COLOR_PERSONAL;
  char counter[12] = "";
  if (pageCount > 1) {
    snprintf(counter, sizeof(counter), "%d/%d", pageIndex + 1, pageCount);
  }
  int counterW = textWidth(counter, FONT_TEXT) + 1;
  line(page.work ? tr("TRABAJO", "WORK") : "PERSONAL", FONT_TITLE, scope, PAD_X, PAD_Y, INNER_W - counterW - 4);
  if (counter[0]) {
    line(counter, FONT_TEXT, scope, PAD_X + INNER_W - counterW, PAD_Y + TITLE_H - ROW_H, counterW, LV_TEXT_ALIGN_RIGHT);
  }
  rect(PAD_X, PAD_Y + TITLE_H + 2, INNER_W, 1, scope);

  int y = CONTENT_TOP;
  if (page.chunkCount == 0) {
    line(tr("Sin cuentas", "No accounts"), FONT_TEXT, COLOR_MUTED, PAD_X, y, INNER_W);
  }
  for (int i = 0; i < page.chunkCount; i++) {
    y = renderChunk(chunks[page.firstChunk + i], y);
  }

  // Footer: when the API was read, or why the data on screen is old.
  char buffer[32];
  char footer[48];
  lv_color_t footerColor = COLOR_MUTED;
  if (view.link == LINK_OK) {
    snprintf(footer, sizeof(footer), tr("Leido %s", "Read %s"), view.data.readAt);
  } else {
    snprintf(footer, sizeof(footer), LV_SYMBOL_WARNING " %s", linkText(buffer, sizeof(buffer)));
    footerColor = linkColor();
  }
  char tail[48] = "";
  bool mock = false;
  if (TAIL_IN_FOOTER && page.chunkCount > 0) {
    const Account &a = view.data.accounts[chunks[page.firstChunk].account];
    char updated[32];
    updatedText(updated, sizeof(updated), a);
    mock = a.mock;
    snprintf(tail, sizeof(tail), "%s%s", updated, mock ? " MOCK" : "");
  }
  pair(footer, FONT_SMALL, footerColor, tail, FONT_SMALL, mock ? COLOR_MEDIUM : COLOR_MUTED, PAD_X, FOOTER_Y, INNER_W);
}

static void render() {
  lv_obj_clean(lv_scr_act());
  if (pageCount == 0) {
    renderMessage();
  } else {
    renderPage();
  }
}

void uiBegin() {
  lv_obj_t *screen = lv_scr_act();
  lv_obj_set_style_bg_color(screen, lv_color_hex(0x000000), 0);
  lv_obj_set_style_bg_opa(screen, LV_OPA_COVER, 0);
  lv_obj_clear_flag(screen, LV_OBJ_FLAG_SCROLLABLE);
  view.link = LINK_STARTING;
  render();
  lastRotate = millis();
}

void uiLoop() {
  uint32_t now = millis();
  bool changed = netRead(view, seenVersion);
  bool stale = view.hasData && view.link != LINK_OK && now - view.lastOkMs > STALE_MARK_MS;
  if (changed || stale != markedStale) {
    markedStale = stale;
    buildPages();
    if (pageIndex >= pageCount) {
      pageIndex = 0;
    }
    render();
  }
  if (pageCount > 1 && now - lastRotate >= ROTATE_MS) {
    lastRotate = now;
    pageIndex = (pageIndex + 1) % pageCount;
    render();
  }
}
