package freight

import (
	"testing"
	"time"
)

func tm(s string) time.Time { t, _ := time.Parse("2006-01-02T15:04", s); return t }

var trucks = []Truck{
	{Reg: "MH-04-KT-7781", Vehicle: "32 ft MXL container", Corridor: "BOM-DEL", PayloadT: 21, WindowStart: tm("2026-11-27T06:00"), WindowEnd: tm("2026-11-28T20:00"), DetourKm: 18, QuoteINR: 38500, Reliability: 4.8, DocsValid: true},
	{Reg: "HR-55-AL-2290", Vehicle: "32 ft SXL container", Corridor: "BOM-DEL", PayloadT: 16, WindowStart: tm("2026-11-27T12:00"), WindowEnd: tm("2026-11-29T12:00"), DetourKm: 42, QuoteINR: 35200, Reliability: 4.5, DocsValid: true},
	{Reg: "PB-10-GH-4412", Vehicle: "19 ft closed body", Corridor: "DEL-LDH", PayloadT: 10, WindowStart: tm("2026-11-26T18:00"), WindowEnd: tm("2026-11-27T10:00"), DetourKm: 6, QuoteINR: 14800, Reliability: 4.9, DocsValid: true},
	{Reg: "MH-12-RT-0915", Vehicle: "20 ft open body", Corridor: "BOM-DEL", PayloadT: 9, WindowStart: tm("2026-11-30T06:00"), WindowEnd: tm("2026-12-01T06:00"), DetourKm: 64, QuoteINR: 29900, Reliability: 4.1, DocsValid: false},
}
var load = Load{Corridor: "BOM-DEL", WeightT: 14, WindowStart: tm("2026-11-27T08:00"), WindowEnd: tm("2026-11-28T18:00"), BudgetINR: 40000}

func TestDelhiMumbaiRanking(t *testing.T) {
	m := Rank(trucks, load)
	if len(m) != 2 || m[0].Truck.Reg != "MH-04-KT-7781" {
		t.Fatalf("unexpected ranking %+v", m)
	}
	if HardFilter(trucks[2], load) != "corridor" || HardFilter(trucks[3], load) != "window" {
		t.Fatal("hard filters wrong")
	}
}

func TestNoPayToRank(t *testing.T) {
	promoted := append([]Truck{}, trucks...)
	promoted[1].Promoted = true
	if Rank(promoted, load)[0].Truck.Reg != Rank(trucks, load)[0].Truck.Reg {
		t.Fatal("promotion must not change ranking")
	}
}

func TestExplicitBookingRequired(t *testing.T) {
	m := Rank(trucks, load)[0]
	if _, err := Book(m, false, "", 2); err != ErrExplicitAcceptance {
		t.Fatal("must never auto-book")
	}
	b, err := Book(m, true, "Ananya Deshpande", 2)
	if err != nil || b.FeeINR != 770 {
		t.Fatalf("booking wrong: %+v %v", b, err)
	}
}
