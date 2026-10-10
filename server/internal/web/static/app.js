// Panel logic. Templates only use x-text and attribute bindings, never x-html:
// everything that comes from the API reaches the page as text.
(() => {
  'use strict';

  const KEYS = { access: 'aqm.accessToken', refresh: 'aqm.refreshToken', expires: 'aqm.accessExpiresAt', lang: 'aqm.lang', scale: 'aqm.displayScale', layout: 'aqm.displayLayout' };
  const LANGUAGES = ['es', 'en'];
  const DEFAULT_LANGUAGE = 'es';
  // Renew this long before the access token expires.
  const EARLY_MS = 60 * 1000;
  const USAGE_POLL_MS = 60 * 1000;
  const LOGIN_POLL_MS = 3 * 1000;
  const ROTATE_MS = 10 * 1000;
  // all: every account of a scope. single-*: one account at a time, larger text.
  const LAYOUTS = ['all', 'single-v', 'single-h'];
  // Providers with a logo in static/logos/<provider>.jpg; the rest show their initial.
  const LOGOS = ['anthropic', 'openai', 'cursor', 'openrouter'];
  const STATUS_ICONS = { ok: '✓', stale: '↻', auth_required: '⚠', unavailable: '✕' };

  const store = {
    get(key) { try { return window.localStorage.getItem(key); } catch (e) { return null; } },
    set(key, value) { try { window.localStorage.setItem(key, value); } catch (e) { /* storage unavailable */ } },
    remove(key) { try { window.localStorage.removeItem(key); } catch (e) { /* storage unavailable */ } },
  };

  const session = {
    read() {
      return { access: store.get(KEYS.access), refresh: store.get(KEYS.refresh), expires: Number(store.get(KEYS.expires)) || 0 };
    },
    save(tokens) {
      store.set(KEYS.access, tokens.accessToken);
      store.set(KEYS.refresh, tokens.refreshToken);
      store.set(KEYS.expires, String(Date.now() + tokens.expiresIn * 1000));
    },
    clear() {
      store.remove(KEYS.access);
      store.remove(KEYS.refresh);
      store.remove(KEYS.expires);
    },
    active() { return Boolean(store.get(KEYS.refresh)); },
  };

  class SessionLost extends Error {}
  class NetworkError extends Error {}
  class ApiError extends Error {
    constructor(status, retryAfter) {
      super('HTTP ' + status);
      this.status = status;
      this.retryAfter = retryAfter;
    }
  }

  async function send(path, options) {
    try {
      return await fetch(path, Object.assign({ cache: 'no-store', credentials: 'omit' }, options));
    } catch (e) {
      throw new NetworkError();
    }
  }
  function jsonBody(value) {
    return { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(value) };
  }

  // One renewal at a time: inside this tab through the shared promise, and
  // across tabs through a Web Lock, because each refresh token works only once.
  let renewing = null;
  function lock(task) {
    return navigator.locks && navigator.locks.request ? navigator.locks.request('aqm-session-refresh', task) : task();
  }
  function renew(staleAccess) {
    if (!renewing) {
      renewing = lock(async () => {
        const current = session.read();
        if (!current.refresh) throw new SessionLost();
        // Another tab may have renewed while this one waited for the lock.
        if (current.access && current.access !== staleAccess && current.expires - Date.now() > EARLY_MS) return current.access;
        const res = await send('/auth/refresh', jsonBody({ refreshToken: current.refresh }));
        if (res.status >= 400 && res.status < 500) {
          session.clear();
          throw new SessionLost();
        }
        if (!res.ok) throw new ApiError(res.status);
        const tokens = await res.json();
        session.save(tokens);
        return tokens.accessToken;
      }).finally(() => { renewing = null; });
    }
    return renewing;
  }

  // request sends an authenticated call. It renews the access token shortly
  // before it expires and, on a 401, renews and repeats the call exactly once.
  async function request(path, options = {}) {
    const current = session.read();
    if (!current.refresh) throw new SessionLost();
    let access = current.access;
    if (!access || current.expires - Date.now() < EARLY_MS) access = await renew(access);
    const call = (token) => {
      const headers = Object.assign({}, options.headers, { Authorization: 'Bearer ' + token });
      const init = { method: options.method || 'GET', headers };
      if (options.body !== undefined) {
        headers['Content-Type'] = 'application/json';
        init.body = JSON.stringify(options.body);
      }
      return send(path, init);
    };
    let res = await call(access);
    if (res.status === 401) {
      access = await renew(access);
      res = await call(access);
      if (res.status === 401) {
        session.clear();
        throw new SessionLost();
      }
    }
    if (!res.ok) throw new ApiError(res.status, Number(res.headers.get('Retry-After')) || 0);
    return res.status === 204 ? null : res.json();
  }

  function lookup(messages, key) {
    let node = messages;
    for (const part of key.split('.')) {
      if (node === null || typeof node !== 'object' || !(part in node)) return key;
      node = node[part];
    }
    return typeof node === 'string' ? node : key;
  }
  function fill(text, values) {
    return text.replace(/\{(\w+)\}/g, (match, name) => (name in values ? String(values[name]) : match));
  }
  // Only https links coming from the API are turned into clickable links.
  function safeUrl(value) {
    try {
      return new URL(value).protocol === 'https:' ? value : '';
    } catch (e) {
      return '';
    }
  }
  function timeZone() {
    try { return Intl.DateTimeFormat().resolvedOptions().timeZone || ''; } catch (e) { return ''; }
  }

  document.addEventListener('alpine:init', () => {
    window.Alpine.data('panel', () => ({
      booting: true,
      ready: false,
      lang: DEFAULT_LANGUAGE,
      messages: {},
      authed: false,
      view: 'dashboard',
      login: { busy: false, error: '' },
      usage: { loading: false, loaded: false, error: '', raw: [], groups: [], checkedText: '' },
      display: { scope: 'personal', scale: 2, rotate: false, clock: '', layout: 'all', index: 0 },
      settings: { loading: false, loaded: false, error: '', storage: true, raw: [], flows: {}, accounts: [] },
      timers: { renew: 0, usage: 0, rotate: 0, clock: 0, login: {} },

      async init() {
        const saved = store.get(KEYS.lang);
        await this.loadLanguage(LANGUAGES.includes(saved) ? saved : DEFAULT_LANGUAGE);
        const scale = Number(store.get(KEYS.scale));
        if ([1, 2, 3].includes(scale)) this.display.scale = scale;
        const layout = store.get(KEYS.layout);
        if (LAYOUTS.includes(layout)) this.display.layout = layout;
        this.route();
        window.addEventListener('hashchange', () => this.route());
        // Keep tabs in sync when another one logs in or out.
        window.addEventListener('storage', (event) => {
          if (event.key === KEYS.refresh || event.key === null) this.syncSession();
        });
        document.addEventListener('visibilitychange', () => {
          if (!document.hidden && this.authed) this.loadUsage();
        });
        this.tick();
        this.timers.clock = window.setInterval(() => this.tick(), 15 * 1000);
        this.authed = session.active();
        this.booting = false;
        this.ready = true;
        if (this.authed) this.start();
      },

      // ---- i18n ----
      tr(key, values) {
        const text = lookup(this.messages, key);
        return values ? fill(text, values) : text;
      },
      async loadLanguage(lang) {
        try {
          const res = await send('/static/i18n/' + lang + '.json', {});
          if (!res.ok) throw new ApiError(res.status);
          this.messages = await res.json();
          this.lang = lang;
          document.documentElement.lang = lang;
        } catch (e) {
          // Keep the previous language; keys are shown if nothing loaded.
        }
      },
      async setLanguage(lang) {
        if (!LANGUAGES.includes(lang) || lang === this.lang) return;
        await this.loadLanguage(lang);
        store.set(KEYS.lang, this.lang);
        this.rebuildUsage();
        this.rebuildSettings();
        this.tick();
        // Quota names are translated by the API.
        if (this.authed) this.loadUsage();
      },
      time(value, options) {
        if (!value) return '';
        const date = new Date(value);
        if (Number.isNaN(date.getTime())) return '';
        return new Intl.DateTimeFormat(this.lang, options || { dateStyle: 'short', timeStyle: 'short' }).format(date);
      },
      number(value) {
        return new Intl.NumberFormat(this.lang, { maximumFractionDigits: 1 }).format(value);
      },
      money(value) {
        return new Intl.NumberFormat(this.lang, { style: 'currency', currency: 'USD' }).format(value);
      },
      errorText(error) {
        if (error instanceof NetworkError) return this.tr('errors.network');
        if (error instanceof ApiError) {
          const known = { 400: 'invalid', 404: 'notFound', 409: 'busy', 429: 'rate', 502: 'provider', 503: 'disabled' }[error.status];
          return known ? this.tr('errors.' + known) : this.tr('errors.server', { status: error.status });
        }
        return this.tr('errors.server', { status: '?' });
      },

      // ---- navigation ----
      route() {
        const name = window.location.hash.replace(/^#\/?/, '');
        this.view = ['dashboard', 'display', 'settings'].includes(name) ? name : 'dashboard';
        if (this.authed && this.view === 'settings') this.loadSettings();
      },
      navClass(name) { return this.view === name ? 'active' : ''; },
      get showLogin() { return this.ready && !this.authed; },
      get showDashboard() { return this.authed && this.view === 'dashboard'; },
      get showDisplay() { return this.authed && this.view === 'display'; },
      get showSettings() { return this.authed && this.view === 'settings'; },

      // ---- session ----
      get loginButton() { return this.tr(this.login.busy ? 'login.busy' : 'login.submit'); },
      async submitLogin(event) {
        const form = event.target;
        const data = new FormData(form);
        const username = String(data.get('username') || '').trim();
        const password = String(data.get('password') || '');
        if (!username || !password) {
          this.login.error = this.tr('login.required');
          return;
        }
        this.login.busy = true;
        this.login.error = '';
        try {
          const res = await send('/auth/login', jsonBody({ username, password }));
          if (res.status === 401) {
            this.login.error = this.tr('login.invalid');
          } else if (res.status === 429) {
            const minutes = Math.max(1, Math.ceil((Number(res.headers.get('Retry-After')) || 60) / 60));
            this.login.error = this.tr('login.rate', { minutes });
          } else if (res.status === 503) {
            this.login.error = this.tr('login.disabled');
          } else if (!res.ok) {
            this.login.error = this.tr('errors.server', { status: res.status });
          } else {
            session.save(await res.json());
            form.reset();
            this.authed = true;
            this.start();
          }
        } catch (error) {
          this.login.error = this.errorText(error);
        } finally {
          this.login.busy = false;
        }
      },
      async logout() {
        const current = session.read();
        this.stop();
        session.clear();
        this.authed = false;
        this.login.error = '';
        // Revoke on the server too; the browser copy is already gone.
        try {
          const options = jsonBody({ refreshToken: current.refresh || '' });
          if (current.access) options.headers.Authorization = 'Bearer ' + current.access;
          await send('/auth/logout', options);
        } catch (e) { /* the local session is already cleared */ }
      },
      lost() {
        this.stop();
        session.clear();
        this.authed = false;
        this.login.error = this.tr('login.expired');
      },
      syncSession() {
        const active = session.active();
        if (active && !this.authed) {
          this.authed = true;
          this.login.error = '';
          this.start();
        } else if (!active && this.authed) {
          this.stop();
          this.authed = false;
        }
      },
      // guard runs an API task and sends the user back to the login when the session is gone.
      async guard(task) {
        try {
          return await task();
        } catch (error) {
          if (error instanceof SessionLost) {
            this.lost();
            return undefined;
          }
          throw error;
        }
      },
      scheduleRenewal() {
        window.clearTimeout(this.timers.renew);
        if (!this.authed) return;
        const wait = Math.max(session.read().expires - Date.now() - EARLY_MS, 5000);
        this.timers.renew = window.setTimeout(async () => {
          try {
            await renew(session.read().access);
          } catch (error) {
            if (error instanceof SessionLost) {
              this.lost();
              return;
            }
          }
          this.scheduleRenewal();
        }, wait);
      },
      start() {
        this.scheduleRenewal();
        this.loadUsage();
        window.clearInterval(this.timers.usage);
        this.timers.usage = window.setInterval(() => {
          if (!document.hidden) this.loadUsage();
        }, USAGE_POLL_MS);
        if (this.view === 'settings') this.loadSettings();
      },
      stop() {
        window.clearTimeout(this.timers.renew);
        window.clearInterval(this.timers.usage);
        window.clearInterval(this.timers.rotate);
        Object.keys(this.timers.login).forEach((id) => window.clearTimeout(this.timers.login[id]));
        this.timers.login = {};
        this.display.rotate = false;
        this.usage = { loading: false, loaded: false, error: '', raw: [], groups: [], checkedText: '' };
        this.settings = { loading: false, loaded: false, error: '', storage: true, raw: [], flows: {}, accounts: [] };
      },

      // ---- usage: dashboard and display ----
      get refreshButton() { return this.tr(this.usage.loading ? 'common.refreshing' : 'common.refresh'); },
      get usageFirstLoad() { return this.usage.loading && !this.usage.loaded; },
      get groups() { return this.usage.groups; },
      async loadUsage() {
        if (this.usage.loading || !this.authed) return;
        this.usage.loading = true;
        try {
          const headers = { 'Accept-Language': this.lang };
          if (timeZone()) headers['X-Timezone'] = timeZone();
          const data = await this.guard(() => request('/usage', { headers }));
          if (data) {
            this.usage.raw = Array.isArray(data.accounts) ? data.accounts : [];
            this.usage.loaded = true;
            this.usage.error = '';
            this.usage.checkedAt = Date.now();
            this.rebuildUsage();
          }
        } catch (error) {
          // Keep what was on screen: old values stay labelled with their own update time.
          this.usage.error = this.errorText(error);
        } finally {
          this.usage.loading = false;
        }
      },
      rebuildUsage() {
        if (!this.usage.loaded) return;
        this.usage.checkedText = this.tr('common.checked', { time: this.time(this.usage.checkedAt, { timeStyle: 'short' }) });
        this.usage.groups = ['personal', 'work'].map((scope) => {
          const accounts = this.usage.raw.filter((a) => a.scope === scope).map((a) => this.accountView(a));
          return { scope, title: this.tr('scope.' + scope), titleClass: 'title-' + scope, accounts, empty: accounts.length === 0 };
        });
      },
      accountView(account) {
        const status = ['ok', 'stale', 'auth_required', 'unavailable'].includes(account.status) ? account.status : 'unavailable';
        const quotas = (Array.isArray(account.quotas) ? account.quotas : []).map((q) => this.quotaView(q));
        const logo = LOGOS.includes(account.provider) ? '/static/logos/' + account.provider + '.jpg' : '';
        return {
          id: account.id,
          name: account.name,
          scope: account.scope,
          providerLabel: this.tr('provider.' + account.provider),
          logo,
          noLogo: !logo,
          initial: String(account.name || '?').charAt(0).toUpperCase(),
          mock: account.source === 'mock',
          statusClass: 'status status-' + status,
          statusIcon: STATUS_ICONS[status],
          statusLabel: this.tr('status.' + status),
          statusHint: status === 'ok' ? '' : this.tr('status.hint.' + status),
          deviceStatus: status === 'ok' ? '' : STATUS_ICONS[status] + ' ' + this.tr('status.' + status),
          deviceStatusClass: 'd-status-' + status,
          updatedText: account.updatedAt ? this.tr('common.updated', { time: this.time(account.updatedAt) }) : this.tr('common.never'),
          updatedShort: account.updatedAt ? this.tr('display.updated', { time: this.time(account.updatedAt, { timeStyle: 'short' }) }) : '',
          quotas,
          deviceCardClass: quotas.length > 2 ? 'd-card d-many' : 'd-card',
          empty: quotas.length === 0,
        };
      },
      // quotaView never invents numbers: a missing value is shown as missing.
      quotaView(quota) {
        const used = typeof quota.used === 'number' ? quota.used : null;
        const limit = typeof quota.limit === 'number' ? quota.limit : null;
        const view = { id: quota.id, name: quota.name, usedText: this.tr('quota.unknown'), leftText: '—', leftShort: '—', percent: 0, hasBar: false, level: 'ok', resetText: '', resetShort: '' };
        const amount = (value) => {
          if (quota.unit === 'usd') return this.money(value);
          if (quota.unit === 'percent') return this.number(value) + '%';
          return this.number(value) + ' ' + this.tr('quota.unit.' + quota.unit);
        };
        if (used !== null && limit !== null) {
          const left = Math.max(limit - used, 0);
          if (quota.unit === 'percent') {
            view.usedText = this.tr('quota.usedPercent', { used: this.number(used) });
            view.leftText = this.tr('quota.leftPercent', { left: this.number(left) });
          } else {
            view.usedText = this.tr('quota.usedOf', { used: amount(used), limit: amount(limit) });
            view.leftText = this.tr('quota.left', { left: amount(left) });
          }
          view.leftShort = amount(left);
          if (used >= limit) {
            view.leftText = this.tr('quota.exhausted');
            view.leftShort = this.tr('quota.exhausted');
          }
          if (limit > 0) {
            view.percent = Math.min(Math.max((used / limit) * 100, 0), 100);
            view.hasBar = true;
            view.level = view.percent >= 90 ? 'crit' : view.percent >= 70 ? 'warn' : 'ok';
          }
        } else if (used !== null) {
          view.usedText = this.tr('quota.usedOnly', { used: amount(used) });
          view.leftText = view.usedText;
          view.leftShort = amount(used);
        }
        if (quota.resetsAt) {
          const time = this.time(quota.resetsAt);
          if (time) {
            view.resetText = this.tr('quota.resets', { time });
            view.resetShort = '↻ ' + time;
          }
        }
        return view;
      },

      tick() {
        this.display.clock = this.time(Date.now(), { timeStyle: 'short' });
      },
      scopeClass(scope) { return this.display.scope === scope ? 'active' : ''; },
      scopePressed(scope) { return this.display.scope === scope ? 'true' : 'false'; },
      setScope(scope) {
        this.display.scope = scope;
        const first = this.allAccounts.findIndex((a) => a.scope === scope);
        if (first >= 0) this.display.index = first;
      },
      get singleLayout() { return this.display.layout !== 'all'; },
      layoutClass(layout) { return this.display.layout === layout ? 'active' : ''; },
      layoutPressed(layout) { return this.display.layout === layout ? 'true' : 'false'; },
      setLayout(layout) {
        this.display.layout = layout;
        store.set(KEYS.layout, layout);
        this.setScope(this.display.scope);
      },
      get allAccounts() { return this.usage.groups.flatMap((g) => g.accounts); },
      get currentAccount() {
        const all = this.allAccounts;
        return all.length ? all[this.display.index % all.length] : null;
      },
      // Single layouts walk every account, personal first and then work.
      stepAccount(delta) {
        const all = this.allAccounts;
        if (!all.length) return;
        this.display.index = ((this.display.index % all.length) + delta + all.length) % all.length;
        this.display.scope = all[this.display.index].scope;
      },
      prevAccount() { this.stepAccount(-1); },
      nextAccount() { this.stepAccount(1); },
      scaleClass(scale) { return this.display.scale === scale ? 'active' : ''; },
      setScale(scale) {
        this.display.scale = scale;
        store.set(KEYS.scale, String(scale));
      },
      setRotate(enabled) {
        this.display.rotate = Boolean(enabled);
        window.clearInterval(this.timers.rotate);
        if (this.display.rotate) {
          this.timers.rotate = window.setInterval(() => {
            if (this.singleLayout) this.stepAccount(1);
            else this.display.scope = this.display.scope === 'personal' ? 'work' : 'personal';
          }, ROTATE_MS);
        }
      },
      get deviceWrapClass() { return 'device-wrap scale-' + this.display.scale + (this.display.layout === 'single-h' ? ' land' : ''); },
      get deviceClass() {
        const layout = { all: '', 'single-v': ' single', 'single-h': ' single land' }[this.display.layout];
        return 'device ' + this.display.scope + layout;
      },
      get deviceCounter() {
        const total = this.allAccounts.length;
        return this.singleLayout && total > 1 ? ((this.display.index % total) + 1) + '/' + total : '';
      },
      get deviceTitle() { return this.tr('scope.' + this.display.scope).toUpperCase(); },
      get deviceClock() { return this.display.clock; },
      get deviceAccounts() {
        if (this.singleLayout) {
          const current = this.currentAccount;
          return current && current.scope === this.display.scope ? [current] : [];
        }
        const group = this.usage.groups.find((g) => g.scope === this.display.scope);
        return group ? group.accounts : [];
      },
      get deviceLoading() { return !this.usage.loaded; },
      get deviceEmpty() { return this.usage.loaded && this.deviceAccounts.length === 0; },

      // ---- settings ----
      get settingsFirstLoad() { return this.settings.loading && !this.settings.loaded; },
      get storageOff() { return this.settings.loaded && !this.settings.storage; },
      flow(id) {
        if (!this.settings.flows[id]) {
          this.settings.flows[id] = { status: 'unknown', userCode: '', url: '', expires: '', busy: false, message: '', error: '' };
        }
        return this.settings.flows[id];
      },
      async loadSettings() {
        if (this.settings.loading || !this.authed) return;
        this.settings.loading = true;
        try {
          const data = await this.guard(() => request('/accounts'));
          if (data) {
            this.settings.raw = Array.isArray(data.accounts) ? data.accounts : [];
            this.settings.storage = data.credentialStorage === true;
            this.settings.loaded = true;
            this.settings.error = '';
            this.rebuildSettings();
            this.settings.raw.filter((a) => a.login).forEach((a) => {
              if (this.flow(a.id).status === 'unknown') this.checkLogin(a.id, true);
            });
          }
        } catch (error) {
          this.settings.error = this.errorText(error);
        } finally {
          this.settings.loading = false;
        }
      },
      rebuildSettings() {
        this.settings.accounts = this.settings.raw.map((account) => {
          const flow = this.flow(account.id);
          const credential = account.credential || null;
          const origin = credential && credential.configured ? credential.origin : '';
          const field = lookup(this.messages, 'settings.credential.field.' + account.provider);
          return {
            id: account.id,
            name: account.name,
            subtitle: this.tr('settings.subtitle', { provider: this.tr('provider.' + account.provider), scope: this.tr('scope.' + account.scope) }),
            isCodex: account.login === 'codex',
            isClaude: account.login === 'claude',
            hasCredential: credential !== null,
            plain: !account.login && credential === null,
            plainText: this.tr('settings.plain.' + (account.source === 'mock' || account.source === 'collector' ? account.source : 'none')),
            loginLabel: this.tr('settings.login.' + flow.status),
            userCode: flow.status === 'pending' ? flow.userCode : '',
            url: flow.status === 'pending' ? flow.url : '',
            expiresText: flow.expires ? this.tr('settings.expires', { time: this.time(flow.expires, { timeStyle: 'short' }) }) : '',
            busy: flow.busy,
            credentialLabel: this.tr('settings.credential.' + (origin === 'stored' || origin === 'env' ? origin : 'missing')),
            credentialField: field.startsWith('settings.') ? this.tr('settings.credential.field.default') : field,
            saveLabel: this.tr(flow.busy ? 'settings.credential.saving' : origin ? 'settings.credential.replace' : 'settings.credential.save'),
            saveDisabled: flow.busy || !this.settings.storage,
            message: flow.message,
            error: flow.error,
          };
        });
      },
      loginPath(id) {
        const account = this.settings.raw.find((a) => a.id === id);
        return account && account.login ? '/accounts/' + encodeURIComponent(id) + '/' + account.login + '/login' : '';
      },
      applyLoginState(id, state) {
        const flow = this.flow(id);
        const known = ['unknown', 'starting', 'pending', 'authenticated', 'auth_required', 'failed'];
        const before = flow.status;
        flow.status = known.includes(state.status) ? state.status : 'unknown';
        flow.userCode = typeof state.userCode === 'string' ? state.userCode : '';
        flow.url = safeUrl(state.verificationUrl || '');
        flow.expires = state.expiresAt || '';
        window.clearTimeout(this.timers.login[id]);
        if (flow.status === 'pending' || flow.status === 'starting') {
          this.timers.login[id] = window.setTimeout(() => this.checkLogin(id, true), LOGIN_POLL_MS);
        } else if (flow.status === 'authenticated' && (before === 'pending' || before === 'starting')) {
          const account = this.settings.raw.find((a) => a.id === id);
          flow.message = this.tr('settings.' + (account ? account.login : 'codex') + '.done');
          window.setTimeout(() => this.loadUsage(), 4000);
        }
      },
      // run wraps a settings action with the busy flag and translated errors.
      async run(id, quiet, task) {
        const flow = this.flow(id);
        if (!quiet) {
          flow.busy = true;
          flow.error = '';
          flow.message = '';
          this.rebuildSettings();
        }
        try {
          await this.guard(task);
        } catch (error) {
          if (!quiet) flow.error = this.errorText(error);
        } finally {
          flow.busy = false;
          this.rebuildSettings();
        }
      },
      async checkLogin(id, quiet) {
        const path = this.loginPath(id);
        if (!path || !this.authed) return;
        await this.run(id, quiet === true, async () => {
          this.applyLoginState(id, await request(path));
        });
      },
      async startCodex(id) {
        await this.run(id, false, async () => {
          this.applyLoginState(id, await request(this.loginPath(id), { method: 'POST' }));
          if (this.flow(id).status === 'pending') this.flow(id).message = this.tr('settings.codex.started');
        });
      },
      async startClaude(id) {
        await this.run(id, false, async () => {
          this.applyLoginState(id, await request(this.loginPath(id), { method: 'POST' }));
          if (this.flow(id).status === 'pending') this.flow(id).message = this.tr('settings.claude.started');
        });
      },
      async submitClaudeCode(event, id) {
        const form = event.target;
        const code = String(new FormData(form).get('code') || '').trim();
        if (!code) {
          this.flow(id).error = this.tr('settings.claude.codeRequired');
          this.rebuildSettings();
          return;
        }
        await this.run(id, false, async () => {
          const state = await request(this.loginPath(id), { method: 'PUT', body: { code } });
          form.reset();
          this.flow(id).status = 'pending';
          this.applyLoginState(id, state);
        });
      },
      async saveCredential(event, id) {
        const form = event.target;
        const value = String(new FormData(form).get('value') || '').trim();
        if (!value) {
          this.flow(id).error = this.tr('settings.credential.required');
          this.rebuildSettings();
          return;
        }
        await this.run(id, false, async () => {
          const state = await request('/accounts/' + encodeURIComponent(id) + '/credential', { method: 'PUT', body: { value } });
          form.reset();
          const account = this.settings.raw.find((a) => a.id === id);
          if (account) account.credential = { configured: state.configured === true, origin: state.origin || '' };
          this.flow(id).message = this.tr('settings.credential.saved');
          window.setTimeout(() => this.loadUsage(), 4000);
        });
      },
    }));
  });
})();
