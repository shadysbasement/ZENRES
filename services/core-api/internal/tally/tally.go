// Package tally implements idempotent Tally.ERP 9 Day Book imports keyed on (org_id, source, external_ref).
package tally

import (
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Voucher struct {
	Source       string
	ExternalRef  string
	Type         string
	Counterparty string
	Date         time.Time
	AmountINR    float64
	IsExport     bool
}

var fieldMatchers = []struct {
	key string
	re  *regexp.Regexp
}{
	{"date", regexp.MustCompile(`(?i)^(date|voucher ?date|vch ?date)$`)},
	{"ref", regexp.MustCompile(`(?i)(voucher ?(no|number)|vch ?no|ref|guid)`)},
	{"type", regexp.MustCompile(`(?i)(voucher ?type|vch ?type|type)`)},
	{"party", regexp.MustCompile(`(?i)(party|ledger|particulars|customer|supplier)`)},
	{"amount", regexp.MustCompile(`(?i)(amount|debit|credit|value)`)},
	{"export", regexp.MustCompile(`(?i)(export|foreign|lut)`)},
}

func AutoMap(headers []string) map[string]int {
	m := map[string]int{}
	for i, h := range headers {
		for _, f := range fieldMatchers {
			if _, used := m[f.key]; !used && f.re.MatchString(strings.TrimSpace(h)) {
				m[f.key] = i
				break
			}
		}
	}
	return m
}

var typeMap = map[string]string{"sales": "sales", "sale": "sales", "receipt": "sales", "purchase": "purchase", "payment": "expense", "journal": "expense", "utility": "utility", "freight": "freight"}

func ParseCSV(r io.Reader, source string) ([]Voucher, []error) {
	rd := csv.NewReader(r)
	rd.FieldsPerRecord = -1
	rows, err := rd.ReadAll()
	if err != nil || len(rows) < 2 {
		return nil, []error{fmt.Errorf("empty or unreadable CSV: %v", err)}
	}
	m := AutoMap(rows[0])
	var out []Voucher
	var errs []error
	for i, row := range rows[1:] {
		get := func(k string) string {
			if idx, ok := m[k]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}
		amt, err := strconv.ParseFloat(strings.NewReplacer(",", "", "₹", "", " ", "").Replace(get("amount")), 64)
		if get("ref") == "" || err != nil {
			errs = append(errs, fmt.Errorf("row %d: missing voucher ref or amount", i+2))
			continue
		}
		d, _ := time.Parse("2-Jan-2006", get("date"))
		t := typeMap[strings.ToLower(get("type"))]
		if t == "" {
			t = "expense"
		}
		out = append(out, Voucher{Source: source, ExternalRef: get("ref"), Type: t, Counterparty: get("party"), Date: d, AmountINR: amt, IsExport: regexp.MustCompile(`(?i)^(y|yes|true|1)$`).MatchString(get("export"))})
	}
	return out, errs
}

// Upsert merges vouchers into an existing set without duplicates — the in-memory analogue of
// INSERT ... ON CONFLICT (org_id, source, external_ref) DO NOTHING.
func Upsert(existing map[string]Voucher, incoming []Voucher) (inserted, duplicates int) {
	for _, v := range incoming {
		k := v.Source + "|" + v.ExternalRef
		if _, ok := existing[k]; ok {
			duplicates++
			continue
		}
		existing[k] = v
		inserted++
	}
	return
}
