// Package panelauth holds the web panel's own authentication: password
// hashing, login throttling and server-side sessions. It is independent of
// DEVICE_TOKEN and COLLECTOR_TOKEN.
package panelauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	scheme = "pbkdf2-sha256"
	// OWASP recommendation for PBKDF2-HMAC-SHA256.
	iterations = 600000
	// MinPasswordLength applies when a hash is created.
	MinPasswordLength = 6
)

// HashPassword returns "pbkdf2-sha256:<iterations>:<salt>:<hash>". The format
// avoids "$" so it can be pasted into .env files without escaping.
func HashPassword(password string) (string, error) {
	if len(password) < MinPasswordLength {
		return "", fmt.Errorf("password must have at least %d characters", MinPasswordLength)
	}
	return hashWith(password, iterations)
}
func hashWith(password string, iter int) (string, error) {
	salt := make([]byte, 16)
	if _, e := rand.Read(salt); e != nil {
		return "", e
	}
	enc := base64.RawURLEncoding.EncodeToString
	return strings.Join([]string{scheme, strconv.Itoa(iter), enc(salt), enc(pbkdf2([]byte(password), salt, iter, 32))}, ":"), nil
}
func parse(encoded string) (iter int, salt, hash []byte, e error) {
	parts := strings.Split(encoded, ":")
	if len(parts) != 4 || parts[0] != scheme {
		return 0, nil, nil, errors.New("unsupported password hash")
	}
	if iter, e = strconv.Atoi(parts[1]); e != nil || iter < 1 || iter > 10000000 {
		return 0, nil, nil, errors.New("invalid password hash iterations")
	}
	if salt, e = base64.RawURLEncoding.DecodeString(parts[2]); e != nil || len(salt) < 8 {
		return 0, nil, nil, errors.New("invalid password hash salt")
	}
	if hash, e = base64.RawURLEncoding.DecodeString(parts[3]); e != nil || len(hash) < 16 {
		return 0, nil, nil, errors.New("invalid password hash")
	}
	return iter, salt, hash, nil
}

// CheckHash reports whether encoded is a hash this package can verify.
func CheckHash(encoded string) error {
	_, _, _, e := parse(encoded)
	return e
}
func VerifyPassword(encoded, password string) bool {
	iter, salt, hash, e := parse(encoded)
	if e != nil {
		return false
	}
	return subtle.ConstantTimeCompare(hash, pbkdf2([]byte(password), salt, iter, len(hash))) == 1
}

// pbkdf2 implements RFC 8018 with HMAC-SHA256; the standard library only ships it from Go 1.24.
func pbkdf2(password, salt []byte, iter, keyLen int) []byte {
	prf := hmac.New(sha256.New, password)
	size := prf.Size()
	out := make([]byte, 0, keyLen+size)
	var counter [4]byte
	for block := uint32(1); len(out) < keyLen; block++ {
		prf.Reset()
		prf.Write(salt)
		binary.BigEndian.PutUint32(counter[:], block)
		prf.Write(counter[:])
		u := prf.Sum(nil)
		t := append([]byte(nil), u...)
		for i := 1; i < iter; i++ {
			prf.Reset()
			prf.Write(u)
			u = prf.Sum(u[:0])
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t...)
	}
	return out[:keyLen]
}
