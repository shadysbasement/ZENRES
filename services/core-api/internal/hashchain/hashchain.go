// Package hashchain provides the tamper-evident SHA-256 audit chain used by the Evidence Locker
// (Growth+). Each entry hashes prev_hash || canonical JSON (sorted keys), matching the web client.
package hashchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
)

const Genesis = "0000000000000000000000000000000000000000000000000000000000000000"

// Canonical renders JSON with sorted object keys and no insignificant whitespace.
func Canonical(v any) string {
	b, _ := json.Marshal(v)
	var x any
	_ = json.Unmarshal(b, &x)
	var sb strings.Builder
	write(&sb, x)
	return sb.String()
}

func write(sb *strings.Builder, v any) {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sb.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				sb.WriteByte(',')
			}
			kb, _ := json.Marshal(k)
			sb.Write(kb)
			sb.WriteByte(':')
			write(sb, t[k])
		}
		sb.WriteByte('}')
	case []any:
		sb.WriteByte('[')
		for i, e := range t {
			if i > 0 {
				sb.WriteByte(',')
			}
			write(sb, e)
		}
		sb.WriteByte(']')
	default:
		b, _ := json.Marshal(t)
		sb.Write(b)
	}
}

func Link(prev string, record any) string {
	h := sha256.Sum256([]byte(prev + Canonical(record)))
	return hex.EncodeToString(h[:])
}

type Entry struct {
	Record   map[string]any
	PrevHash string
	Hash     string
}

func Build(records []map[string]any) []Entry {
	prev := Genesis
	out := make([]Entry, len(records))
	for i, r := range records {
		h := Link(prev, r)
		out[i] = Entry{Record: r, PrevHash: prev, Hash: h}
		prev = h
	}
	return out
}

// Verify returns the index of the first broken entry, or -1 if the chain is intact.
func Verify(chain []Entry) int {
	prev := Genesis
	for i, e := range chain {
		if e.PrevHash != prev || Link(prev, e.Record) != e.Hash {
			return i
		}
		prev = e.Hash
	}
	return -1
}
