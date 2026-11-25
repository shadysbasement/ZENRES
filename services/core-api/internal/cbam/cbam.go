// Package cbam holds Zevra's deterministic CBAM / carbon math (mirrors apps/web-app/src/lib/carbon.ts).
package cbam

import (
	"encoding/xml"
	"fmt"
	"math"
)

const (
	CertPriceEUR      = 76.0
	DieselKgCO2PerL   = 2.68
	RuleSetDefinitive = "IR (EU) 2025/2621"
)

type Product struct {
	CN          string
	Name        string
	DirectT     float64 // tCO2e per tonne
	IndirectT   float64
	PrecursorT  float64
	DefaultT    float64 // default value incl. markup
	DataQuality string
}

func round(v float64, d int) float64 {
	p := math.Pow(10, float64(d))
	return math.Round(v*p) / p
}

// SpecificEmbedded returns direct + indirect (+ precursors) emissions per tonne.
func SpecificEmbedded(p Product, includePrecursors bool) float64 {
	pre := 0.0
	if includePrecursors {
		pre = p.PrecursorT
	}
	return round(p.DirectT+p.IndirectT+pre, 4)
}

func ExposureEUR(tonnes, intensity, price float64) float64 { return round(tonnes*intensity*price, 2) }

type Comparison struct {
	Actual, Default, DefaultEUR, ActualEUR, SavingsEUR, ReductionPct float64
}

func CompareDefaultVsActual(tonnes float64, p Product, price float64) Comparison {
	a := SpecificEmbedded(p, true)
	d, ae := ExposureEUR(tonnes, p.DefaultT, price), ExposureEUR(tonnes, a, price)
	return Comparison{Actual: a, Default: p.DefaultT, DefaultEUR: d, ActualEUR: ae, SavingsEUR: round(d-ae, 2), ReductionPct: round((p.DefaultT-a)/p.DefaultT*100, 1)}
}

// Article9Deduction caps the carbon price effectively paid in the country of origin at the certificate price.
func Article9Deduction(embeddedT, paidEURPerT, price float64) float64 {
	return round(embeddedT*math.Min(math.Max(0, paidEURPerT), price), 2)
}

// ActivityEmissions converts activity data using a factor expressed in tCO2e per unit.
func ActivityEmissions(qty, efTPerUnit float64) float64 { return round(qty*efTPerUnit, 4) }

// AvoidedFuel is a LOGISTICS estimate. It is never added to or netted against CBAM embedded emissions.
type AvoidedFuel struct {
	Litres float64 `json:"litres"`
	KgCO2  float64 `json:"kg_co2"`
	Scope  string  `json:"scope"`
}

func Avoided(baselineL, matchedL float64) AvoidedFuel {
	l := math.Max(0, baselineL-matchedL)
	return AvoidedFuel{Litres: l, KgCO2: round(l*DieselKgCO2PerL, 1), Scope: "logistics_estimate"}
}

// EmbeddedReport is the statutory CBAM view; it deliberately has no logistics fields.
type EmbeddedReport struct {
	CN         string
	Tonnes     float64
	SpecificT  float64
	TotalTCO2e float64
}

func Embedded(p Product, tonnes float64, includePrecursors bool) EmbeddedReport {
	s := SpecificEmbedded(p, includePrecursors)
	return EmbeddedReport{CN: p.CN, Tonnes: tonnes, SpecificT: s, TotalTCO2e: round(s*tonnes, 4)}
}

/* ---------------- Registry XML ---------------- */

type xmlReport struct {
	XMLName      xml.Name   `xml:"CBAMReport"`
	Xmlns        string     `xml:"xmlns,attr"`
	RuleSet      string     `xml:"ruleSet,attr"`
	Period       string     `xml:"period,attr"`
	Type         string     `xml:"type,attr"`
	Declarant    xmlParty   `xml:"Declarant"`
	Installation xmlInstall `xml:"Installation"`
	Goods        []xmlGoods `xml:"Goods>GoodsImported"`
	Disclaimer   string     `xml:"Disclaimer"`
}
type xmlParty struct {
	Name string `xml:"Name"`
	EORI string `xml:"EORI"`
}
type xmlInstall struct {
	Name     string `xml:"Name"`
	Operator string `xml:"Operator"`
	UNLOCODE string `xml:"UNLOCODE"`
	Country  string `xml:"Country"`
}
type xmlGoods struct {
	Seq      int    `xml:"seq,attr"`
	CN       string `xml:"CommodityCode"`
	Desc     string `xml:"Description"`
	Mass     string `xml:"NetMassTonnes"`
	Emission xmlSEE `xml:"SpecificEmbeddedEmissions"`
	Method   string `xml:"DeterminationMethod"`
	Quality  string `xml:"DataQuality"`
}
type xmlSEE struct {
	Unit     string `xml:"unit,attr"`
	Direct   string `xml:"Direct"`
	Indirect string `xml:"Indirect"`
	Total    string `xml:"Total"`
}

type Line struct {
	Product Product
	Tonnes  float64
}

type ReportMeta struct {
	Period, Declarant, EORI, Installation, Operator, UNLOCODE, Country string
	Consolidated                                                       bool
}

var cnRe = func(s string) bool {
	if len(s) != 8 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// RegistryXML builds a deterministic Registry XML document and validates content rules.
func RegistryXML(meta ReportMeta, lines []Line) ([]byte, error) {
	if meta.EORI == "" || meta.Period == "" {
		return nil, fmt.Errorf("content rule: declarant EORI and period are mandatory")
	}
	r := xmlReport{Xmlns: "urn:eu:taxud:cbam:v1", RuleSet: RuleSetDefinitive, Period: meta.Period, Type: "INSTALLATION",
		Declarant: xmlParty{meta.Declarant, meta.EORI}, Installation: xmlInstall{meta.Installation, meta.Operator, meta.UNLOCODE, meta.Country},
		Disclaimer: "Prepared with Zevra reporting support. Not legal, tax or accounting advice."}
	if meta.Consolidated {
		r.Type = "CONSOLIDATED"
	}
	for i, l := range lines {
		if !cnRe(l.Product.CN) {
			return nil, fmt.Errorf("content rule: CN code %q must be 8 digits", l.Product.CN)
		}
		if l.Tonnes <= 0 {
			return nil, fmt.Errorf("content rule: net mass for %s must be positive", l.Product.CN)
		}
		method := "ACTUAL_DATA"
		if l.Product.DataQuality == "estimated" {
			method = "DEFAULT_VALUES"
		}
		r.Goods = append(r.Goods, xmlGoods{Seq: i + 1, CN: l.Product.CN, Desc: l.Product.Name, Mass: fmt.Sprintf("%.3f", l.Tonnes),
			Emission: xmlSEE{Unit: "tCO2e/t", Direct: fmt.Sprintf("%.4f", l.Product.DirectT), Indirect: fmt.Sprintf("%.4f", l.Product.IndirectT), Total: fmt.Sprintf("%.4f", SpecificEmbedded(l.Product, true))},
			Method:   method, Quality: upper(l.Product.DataQuality)})
	}
	out, err := xml.MarshalIndent(r, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), out...), nil
}

func upper(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 32
		}
	}
	return string(b)
}
