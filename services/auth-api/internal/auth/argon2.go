package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters mandated by the blueprint: time=3, memory=64 MiB, threads=4, keyLen=32, 16-byte salt.
const (
	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 4
	argonKeyLen  uint32 = 32
	saltLen             = 16
	MinPassword         = 12
)

var ErrWeakPassword = errors.New("weak_password")

// HashPassword returns a PHC-formatted Argon2id hash.
func HashPassword(pw string) (string, error) {
	if err := ValidatePassword(pw); err != nil {
		return "", err
	}
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(pw), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

// VerifyPassword checks pw against a PHC hash in constant time.
func VerifyPassword(pw, phc string) bool {
	parts := strings.Split(phc, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var m, t uint32
	var p uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}
	got := argon2.IDKey([]byte(pw), salt, t, m, p, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}

// dummyHash is verified against when no user exists, so login latency does not leak account existence.
var dummyHash, _ = HashPassword("zevra-timing-equaliser-0000")

// VerifyOrDummy runs a full Argon2id verification even when phc is empty.
func VerifyOrDummy(pw, phc string) bool {
	if phc == "" {
		VerifyPassword(pw, dummyHash)
		return false
	}
	return VerifyPassword(pw, phc)
}

// ValidatePassword enforces a minimum of 12 characters and rejects trivially repetitive passwords.
func ValidatePassword(pw string) error {
	if utf8.RuneCountInString(pw) < MinPassword {
		return ErrWeakPassword
	}
	unique := map[rune]struct{}{}
	for _, r := range pw {
		unique[r] = struct{}{}
	}
	if len(unique) < 5 {
		return ErrWeakPassword
	}
	return nil
}
