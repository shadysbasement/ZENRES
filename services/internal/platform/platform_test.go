package platform

import (
	"testing"
	"time"
)

func TestSameSubnet(t *testing.T) {
	if !SameSubnet("10.4.2.17", "10.4.2.200") {
		t.Fatal("expected same /24")
	}
	if SameSubnet("10.4.2.17", "10.4.3.17") {
		t.Fatal("expected different /24")
	}
}

func TestLimiter(t *testing.T) {
	l := NewMemLimiter()
	for i := 0; i < 5; i++ {
		if !l.Allow("ip", 5, time.Minute) {
			t.Fatalf("request %d should pass", i)
		}
	}
	if l.Allow("ip", 5, time.Minute) {
		t.Fatal("6th request should be limited")
	}
}

func TestRandomTokenEntropy(t *testing.T) {
	a, b := RandomToken(32), RandomToken(32)
	if a == b || len(a) != 43 {
		t.Fatalf("unexpected tokens %q %q", a, b)
	}
}
