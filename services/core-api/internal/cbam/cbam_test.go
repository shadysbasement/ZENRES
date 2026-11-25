package cbam

import (
	"encoding/xml"
	"reflect"
	"strings"
	"testing"
)

var bar = Product{CN: "76042910", Name: "Aluminium bars 6063-T6", DirectT: 0.62, IndirectT: 1.71, PrecursorT: 3.17, DefaultT: 7.9, DataQuality: "measured"}

func TestDeterministicMath(t *testing.T) {
	if SpecificEmbedded(bar, true) != 5.5 || SpecificEmbedded(bar, false) != 2.33 {
		t.Fatal("specific embedded emissions wrong")
	}
	c := CompareDefaultVsActual(100, bar, 76)
	if c.DefaultEUR != 60040 || c.ActualEUR != 41800 || c.SavingsEUR != 18240 {
		t.Fatalf("comparison mismatch: %+v", c)
	}
	if Article9Deduction(10, 500, 76) != 760 || Article9Deduction(10, -1, 76) != 0 {
		t.Fatal("article 9 cap wrong")
	}
	if ActivityEmissions(412800, 0.000727) != 300.1056 {
		t.Fatal("activity conversion wrong")
	}
}

func TestLogisticsEmissionsAreSeparate(t *testing.T) {
	a := Avoided(180, 140)
	if a.Litres != 40 || a.KgCO2 != 107.2 || a.Scope != "logistics_estimate" {
		t.Fatalf("avoided fuel wrong: %+v", a)
	}
	// The statutory embedded report type must carry no logistics fields (guards against double counting).
	for i := 0; i < reflect.TypeOf(EmbeddedReport{}).NumField(); i++ {
		name := strings.ToLower(reflect.TypeOf(EmbeddedReport{}).Field(i).Name)
		if strings.Contains(name, "fuel") || strings.Contains(name, "logistic") || strings.Contains(name, "avoided") {
			t.Fatalf("embedded report must not include %s", name)
		}
	}
	if e := Embedded(bar, 10, true); e.TotalTCO2e != 55 {
		t.Fatalf("embedded total wrong: %+v", e)
	}
}

func TestRegistryXMLIsValidAndDeterministic(t *testing.T) {
	meta := ReportMeta{Period: "2026-Q4", Declarant: "Rhein Aluminium Import GmbH", EORI: "DE472918350000", Installation: "Chakan Extrusion Plant", Operator: "Deccan Alloy Works", UNLOCODE: "INPNQ", Country: "IN"}
	a, err := RegistryXML(meta, []Line{{bar, 17.5}})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := RegistryXML(meta, []Line{{bar, 17.5}})
	if string(a) != string(b) {
		t.Fatal("output must be deterministic")
	}
	var parsed struct {
		XMLName xml.Name `xml:"CBAMReport"`
		Goods   []struct {
			CN string `xml:"CommodityCode"`
		} `xml:"Goods>GoodsImported"`
	}
	if err := xml.Unmarshal(a, &parsed); err != nil || len(parsed.Goods) != 1 || parsed.Goods[0].CN != "76042910" {
		t.Fatalf("xml not well-formed: %v", err)
	}
	if _, err := RegistryXML(meta, []Line{{Product{CN: "7604"}, 1}}); err == nil {
		t.Fatal("short CN code must fail content rules")
	}
}
