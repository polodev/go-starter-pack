package pdf

import "testing"

func TestMoneyIsFormattedFromMinorUnits(t *testing.T) {
	cases := map[Money]string{0: "0.00", 5: "0.05", 100: "1.00", 123456: "1,234.56", 4500000: "45,000.00",
		100000000000: "1,000,000,000.00", -250: "-2.50"}
	for money, want := range cases {
		if got := money.String(); got != want {
			t.Errorf("Money(%d) = %q, want %q", int64(money), got, want)
		}
	}
}

func TestTheSampleInvoiceAddsUp(t *testing.T) {
	invoice := SampleInvoice("Someone")
	if got := invoice.Total(); got != 6402500 {
		t.Fatalf("total = %d, want 6402500 (45,000 + 12 x 1,250 + 5 x 805)", int64(got))
	}
}
