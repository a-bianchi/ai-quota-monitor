package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

const (
	accessLifetime  = 15 * time.Minute
	refreshLifetime = 30 * 24 * time.Hour
)

var jwtHeader = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

type claims struct {
	Subject string `json:"sub"`
	Use     string `json:"use"`
	Expires int64  `json:"exp"`
}

// sessionKey derives the signing key from the configured secrets, so tokens
// survive restarts and stop working when the password or COLLECTOR_TOKEN change.
func (a *API) sessionKey() []byte {
	if a.Username == "" || a.Password == "" || a.CollectorToken == "" {
		return nil
	}
	mac := hmac.New(sha256.New, []byte(a.CollectorToken))
	mac.Write([]byte("session\x00" + a.Username + "\x00" + a.Password))
	return mac.Sum(nil)
}
func sign(key []byte, body string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func (a *API) issue(key []byte, use string, lifetime time.Duration) string {
	payload, _ := json.Marshal(claims{Subject: a.Username, Use: use, Expires: time.Now().Add(lifetime).Unix()})
	body := jwtHeader + "." + base64.RawURLEncoding.EncodeToString(payload)
	return body + "." + sign(key, body)
}

// validToken reports whether token is an unexpired JWT issued by this server for the given use.
func (a *API) validToken(token, use string) bool {
	key := a.sessionKey()
	parts := strings.Split(token, ".")
	if key == nil || len(parts) != 3 || parts[0] != jwtHeader {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(parts[2]), []byte(sign(key, parts[0]+"."+parts[1]))) != 1 {
		return false
	}
	payload, e := base64.RawURLEncoding.DecodeString(parts[1])
	var c claims
	if e != nil || json.Unmarshal(payload, &c) != nil {
		return false
	}
	return c.Use == use && c.Subject == a.Username && time.Now().Unix() < c.Expires
}
func (a *API) tokens(w http.ResponseWriter, key []byte) {
	write(w, 200, map[string]any{
		"tokenType":    "Bearer",
		"accessToken":  a.issue(key, "access", accessLifetime),
		"expiresIn":    int(accessLifetime.Seconds()),
		"refreshToken": a.issue(key, "refresh", refreshLifetime),
	})
}

// session wraps the login endpoints: they never cache and are off without credentials.
func (a *API) session(next func(http.ResponseWriter, *http.Request, []byte)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		key := a.sessionKey()
		if key == nil {
			write(w, 503, map[string]string{"error": "login disabled; set AUTH_USERNAME, AUTH_PASSWORD and COLLECTOR_TOKEN"})
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
		next(w, r, key)
	})
}
func (a *API) login(w http.ResponseWriter, r *http.Request, key []byte) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if e := json.NewDecoder(r.Body).Decode(&body); e != nil {
		write(w, 400, map[string]string{"error": "expected JSON with username and password"})
		return
	}
	user := subtle.ConstantTimeCompare([]byte(body.Username), []byte(a.Username))
	pass := subtle.ConstantTimeCompare([]byte(body.Password), []byte(a.Password))
	if user&pass != 1 {
		// Slows down password guessing.
		select {
		case <-time.After(a.failureDelay()):
		case <-r.Context().Done():
		}
		write(w, 401, map[string]string{"error": "invalid username or password"})
		return
	}
	a.tokens(w, key)
}
func (a *API) refresh(w http.ResponseWriter, r *http.Request, key []byte) {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	if e := json.NewDecoder(r.Body).Decode(&body); e != nil {
		write(w, 400, map[string]string{"error": "expected JSON with refreshToken"})
		return
	}
	if !a.validToken(body.RefreshToken, "refresh") {
		write(w, 401, map[string]string{"error": "invalid or expired refresh token; log in again"})
		return
	}
	a.tokens(w, key)
}
func (a *API) failureDelay() time.Duration {
	if a.LoginFailureDelay != 0 {
		return a.LoginFailureDelay
	}
	return time.Second
}
