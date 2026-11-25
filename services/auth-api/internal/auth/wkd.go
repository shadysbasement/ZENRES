package auth

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

var jsonUnmarshal = json.Unmarshal

// zbase32 alphabet per RFC 6189, used by the OpenPGP Web Key Directory.
const zb32 = "ybndrfg8ejkmcpqxot1uwisza345h769"

func zbase32(b []byte) string {
	var out strings.Builder
	var buf uint64
	bits := 0
	for _, c := range b {
		buf = buf<<8 | uint64(c)
		bits += 8
		for bits >= 5 {
			out.WriteByte(zb32[(buf>>(uint(bits)-5))&31])
			bits -= 5
		}
	}
	if bits > 0 {
		out.WriteByte(zb32[(buf<<(5-uint(bits)))&31])
	}
	return out.String()
}

// WKDHash returns the hashed local-part used in WKD URLs.
func WKDHash(local string) string {
	h := sha1.Sum([]byte(strings.ToLower(local)))
	return zbase32(h[:])
}

// WKDURLs returns the advanced then direct WKD lookup URLs for an address.
func WKDURLs(email string) []string {
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return nil
	}
	local, domain := email[:at], strings.ToLower(email[at+1:])
	h := WKDHash(local)
	l := url.QueryEscape(local)
	return []string{
		fmt.Sprintf("https://openpgpkey.%s/.well-known/openpgpkey/%s/hu/%s?l=%s", domain, domain, h, l),
		fmt.Sprintf("https://%s/.well-known/openpgpkey/hu/%s?l=%s", domain, h, l),
	}
}

// WKDClient is swappable in tests.
var WKDClient = &http.Client{Timeout: 4 * time.Second}

// LookupWKD fetches a binary OpenPGP key via WKD. Returns nil when none is published.
func LookupWKD(ctx context.Context, email string) (*crypto.Key, error) {
	for _, u := range WKDURLs(email) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		res, err := WKDClient.Do(req)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(res.Body, 256<<10))
		res.Body.Close()
		if res.StatusCode != 200 || len(body) == 0 {
			continue
		}
		if k, err := crypto.NewKey(body); err == nil {
			return k, nil
		}
		if k, err := crypto.NewKeyFromArmored(string(body)); err == nil {
			return k, nil
		}
	}
	return nil, nil
}
