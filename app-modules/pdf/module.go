// Package pdf renders HTML and CSS to PDF with headless Chrome (chromedp), like a browser's "Save as PDF": page
// size and margins are CSS @page. It ships an example invoice template and pdf:sample-invoice.
package pdf

import "github.com/goravel/framework/contracts/console"

// Module registers the invoice template and the example command.
type Module struct{}

func (m *Module) Name() string      { return "pdf" }
func (m *Module) ViewsPath() string { return "app-modules/pdf/resources/views" }

func (m *Module) Commands() []console.Command { return []console.Command{&SampleInvoiceCommand{}} }
