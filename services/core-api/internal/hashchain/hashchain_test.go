package hashchain

import "testing"

func TestCanonicalMatchesClient(t *testing.T) {
	got := Canonical(map[string]any{"b": 1, "a": []any{2, map[string]any{"d": 3, "c": 4}}})
	if got != `{"a":[2,{"c":4,"d":3}],"b":1}` {
		t.Fatalf("canonical JSON differs from web client: %s", got)
	}
}

func TestTamperDetection(t *testing.T) {
	c := Build([]map[string]any{{"id": 1, "kwh": 100}, {"id": 2, "kwh": 200}, {"id": 3, "kwh": 300}})
	if Verify(c) != -1 {
		t.Fatal("fresh chain must verify")
	}
	c[1].Record["kwh"] = 199
	if Verify(c) != 1 {
		t.Fatal("edit must be detected at index 1")
	}
}
