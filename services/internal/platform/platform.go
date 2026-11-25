// Package platform holds small, dependency-free helpers shared by Zevra's Go services:
// JSON responses, client IP extraction, opaque tokens, rate limiting and security headers.
package platform

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Env returns the environment variable or a fallback.
func Env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// JSON writes v with the given status.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Error writes {"error": code} plus optional extra fields.
func Error(w http.ResponseWriter, status int, code string, extra ...map[string]any) {
	body := map[string]any{"error": code}
	for _, e := range extra {
		for k, v := range e {
			body[k] = v
		}
	}
	JSON(w, status, body)
}

// Decode reads a JSON body (max 1 MiB) into v.
func Decode(r *http.Request, v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// DecodeLoose reads a JSON body tolerating unknown fields.
func DecodeLoose(r *http.Request, v any) error {
	return json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20)).Decode(v)
}

// ClientIP prefers X-Forwarded-For (set by our own SvelteKit proxies) then RemoteAddr.
func ClientIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		ip := strings.TrimSpace(strings.Split(xf, ",")[0])
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// SameSubnet reports whether two IPs share a /24 (IPv4) or /56 (IPv6) prefix.
func SameSubnet(a, b string) bool {
	ia, ib := net.ParseIP(a), net.ParseIP(b)
	if ia == nil || ib == nil {
		return a == b
	}
	if a4, b4 := ia.To4(), ib.To4(); a4 != nil && b4 != nil {
		return a4.Mask(net.CIDRMask(24, 32)).Equal(b4.Mask(net.CIDRMask(24, 32)))
	}
	return ia.Mask(net.CIDRMask(56, 128)).Equal(ib.Mask(net.CIDRMask(56, 128)))
}

// RandomToken returns an opaque URL-safe token of n random bytes.
func RandomToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// RandomHex returns n random bytes hex-encoded.
func RandomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// SHA256 hashes s.
func SHA256(s string) []byte {
	h := sha256.Sum256([]byte(s))
	return h[:]
}

// SecurityHeaders adds baseline hardening headers to every API response.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("Cross-Origin-Resource-Policy", "same-site")
		next.ServeHTTP(w, r)
	})
}

// Limiter is a fixed-window rate limiter. Production deployments can swap in the Redis
// implementation behind the same interface; this in-process version is safe for one replica.
type Limiter interface {
	Allow(key string, max int, window time.Duration) bool
}

type memLimiter struct {
	mu sync.Mutex
	m  map[string]*bucket
}

type bucket struct {
	count int
	reset time.Time
}

// NewMemLimiter returns an in-memory fixed-window limiter.
func NewMemLimiter() Limiter { return &memLimiter{m: map[string]*bucket{}} }

func (l *memLimiter) Allow(key string, max int, window time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	b, ok := l.m[key]
	if !ok || now.After(b.reset) {
		l.m[key] = &bucket{count: 1, reset: now.Add(window)}
		return true
	}
	b.count++
	return b.count <= max
}
