// Package freight implements the deterministic FreightLink corridor matcher (mirrors apps/web-app/src/lib/freight.ts).
package freight

import (
	"errors"
	"math"
	"regexp"
	"sort"
	"time"
)

type Truck struct {
	ID, Reg, Carrier, Vehicle, Corridor string
	PayloadT                            float64
	WindowStart, WindowEnd              time.Time
	DetourKm                            float64
	QuoteINR                            float64
	Reliability                         float64
	DocsValid                           bool
	Promoted                            bool // paid promotion — deliberately NOT a ranking input
}

type Load struct {
	ID, Corridor           string
	WeightT                float64
	WindowStart, WindowEnd time.Time
	BudgetINR              float64
	RequiresClosed         bool
}

type Match struct {
	Truck Truck
	Score float64
}

var closedRe = regexp.MustCompile(`(?i)(container|closed)`)

// HardFilter returns the first failing rule, or "" when the truck is eligible.
func HardFilter(t Truck, l Load) string {
	switch {
	case t.Corridor != l.Corridor:
		return "corridor"
	case t.WindowStart.After(l.WindowEnd) || l.WindowStart.After(t.WindowEnd):
		return "window"
	case t.PayloadT < l.WeightT:
		return "payload"
	case l.RequiresClosed && !closedRe.MatchString(t.Vehicle):
		return "vehicle_type"
	case !t.DocsValid:
		return "documents"
	}
	return ""
}

func clamp(v float64) float64 { return math.Max(0, math.Min(1, v)) }

func Rank(trucks []Truck, l Load) []Match {
	var out []Match
	for _, t := range trucks {
		if HardFilter(t, l) != "" {
			continue
		}
		detour := 1 - math.Min(1, t.DetourKm/100)
		quote := clamp(1 - (t.QuoteINR-l.BudgetINR*0.7)/(l.BudgetINR*0.6))
		rel := math.Max(0, (t.Reliability-3)/2)
		score := math.Round((detour*0.35+quote*0.3+rel*0.25+0.1)*10000) / 10000
		out = append(out, Match{Truck: t, Score: score})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Truck.DetourKm < out[j].Truck.DetourKm
	})
	return out
}

var ErrExplicitAcceptance = errors.New("explicit_acceptance_required")

type Booking struct {
	Status         string  `json:"status"`
	FreightINR     float64 `json:"freight_inr"`
	DisclosedFeePc float64 `json:"disclosed_fee_pct"`
	FeeINR         float64 `json:"fee_inr"`
	AuthorisedBy   string  `json:"authorised_by"`
}

// Book never auto-books: it requires an explicit, attributable acceptance of an itemised quote.
func Book(m Match, accepted bool, authorisedBy string, feePct float64) (Booking, error) {
	if !accepted || authorisedBy == "" {
		return Booking{}, ErrExplicitAcceptance
	}
	return Booking{Status: "confirmed", FreightINR: m.Truck.QuoteINR, DisclosedFeePc: feePct, FeeINR: math.Round(m.Truck.QuoteINR * feePct / 100), AuthorisedBy: authorisedBy}, nil
}
