package tally

import (
	"strings"
	"testing"
)

const sample = `Date,Voucher No,Voucher Type,Particulars,Amount,Export
21-Nov-2026,SAL/202611/0451,Sales,Rhein Aluminium Import GmbH,"21,84,000",Yes
22-Nov-2026,PUR/202611/0231,Purchase,Hindalco Renukoot,"18,40,000",No
`

func TestImportTwiceNoDuplicates(t *testing.T) {
	v, errs := ParseCSV(strings.NewReader(sample), "tally_csv")
	if len(errs) != 0 || len(v) != 2 || v[0].AmountINR != 2184000 || !v[0].IsExport {
		t.Fatalf("parse wrong: %+v %v", v, errs)
	}
	db := map[string]Voucher{}
	if ins, dup := Upsert(db, v); ins != 2 || dup != 0 {
		t.Fatal("first import should insert 2")
	}
	if ins, dup := Upsert(db, v); ins != 0 || dup != 2 || len(db) != 2 {
		t.Fatal("second import must produce zero duplicates")
	}
}
