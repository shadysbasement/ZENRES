package auth

import (
	"strings"
	"testing"
)

func TestArgon2idRoundTrip(t *testing.T) {
	h, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(h, "$argon2id$v=19$m=65536,t=3,p=4$") {
		t.Fatalf("unexpected PHC params: %s", h)
	}
	if !VerifyPassword("correct horse battery staple", h) {
		t.Fatal("valid password rejected")
	}
	if VerifyPassword("correct horse battery stapl3", h) {
		t.Fatal("wrong password accepted")
	}
}

func TestSaltsAreUnique(t *testing.T) {
	a, _ := HashPassword("same password for both")
	b, _ := HashPassword("same password for both")
	if a == b {
		t.Fatal("hashes must differ (random salt)")
	}
}

func TestPasswordPolicy(t *testing.T) {
	for _, pw := range []string{"short", "aaaaaaaaaaaaaaaa", "abababababab"} {
		if ValidatePassword(pw) == nil {
			t.Fatalf("%q should be rejected", pw)
		}
	}
	if ValidatePassword("Deccan-Alloy-2026!") != nil {
		t.Fatal("strong password rejected")
	}
}

func TestDummyVerifyNeverSucceeds(t *testing.T) {
	if VerifyOrDummy("anything at all", "") {
		t.Fatal("dummy verification must fail")
	}
}
