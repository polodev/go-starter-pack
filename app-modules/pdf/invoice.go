package pdf

import (
	"fmt"
	"time"

	"goravel/app/web"
)

// Money is kept in minor units (cents, paisa), never floats.
type Money int64

// String formats minor units as 1,234.56.
func (m Money) String() string {
	sign := ""
	value := int64(m)
	if value < 0 {
		sign, value = "-", -value
	}
	whole, fraction := value/100, value%100
	text := fmt.Sprintf("%d", whole)
	for index := len(text) - 3; index > 0; index -= 3 {
		text = text[:index] + "," + text[index:]
	}
	return fmt.Sprintf("%s%s.%02d", sign, text, fraction)
}

// InvoiceLine is one row of an invoice.
type InvoiceLine struct {
	Description string
	Quantity    int64
	UnitPrice   Money
}

// Total of the line.
func (l InvoiceLine) Total() Money { return Money(l.Quantity) * l.UnitPrice }

// Invoice is an example document: an HTML template (pdf::invoice) rendered to PDF.
type Invoice struct {
	Number   string
	Date     time.Time
	Due      time.Time
	Currency string
	Seller   string
	Customer string
	Lines    []InvoiceLine
}

// Total of the invoice.
func (i Invoice) Total() Money {
	var total Money
	for _, line := range i.Lines {
		total += line.Total()
	}
	return total
}

// SampleInvoice is a filled-in example.
func SampleInvoice(customer string) Invoice {
	now := time.Now()
	return Invoice{
		Number: "INV-" + now.Format("20060102") + "-001", Date: now, Due: now.AddDate(0, 0, 14), Currency: "BDT",
		Seller: "Go Starter Ltd.", Customer: customer,
		Lines: []InvoiceLine{
			{Description: "Website design", Quantity: 1, UnitPrice: 4500000},
			{Description: "Hosting (12 months)", Quantity: 12, UnitPrice: 125000},
			{Description: "Support hours", Quantity: 5, UnitPrice: 80500},
		},
	}
}

// HTML renders the invoice template.
func (i Invoice) HTML() (string, error) {
	return web.Render("pdf::invoice", map[string]any{"invoice": i, "lang": "en"})
}

// PDF renders the invoice to PDF bytes.
func (i Invoice) PDF() ([]byte, error) {
	html, err := i.HTML()
	if err != nil {
		return nil, err
	}
	return FromHTML(html)
}
