package pdf

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"

	"goravel/app/facades"
)

// SampleInvoiceCommand is `pdf:sample-invoice --output=storage/app/invoice.pdf`, which writes the example invoice,
// or `pdf:sample-invoice --email=you@example.com`, which mails it as an attachment (Invoice.Email).
type SampleInvoiceCommand struct{}

func (r *SampleInvoiceCommand) Signature() string { return "pdf:sample-invoice" }
func (r *SampleInvoiceCommand) Description() string {
	return "Render the example invoice to a PDF file, or email it with --email"
}
func (r *SampleInvoiceCommand) Extend() command.Extend {
	return command.Extend{Category: "pdf", Flags: []command.Flag{
		&command.StringFlag{Name: "output", Value: "storage/app/invoice.pdf", Usage: "where to write the PDF"},
		&command.StringFlag{Name: "email", Usage: "send it to this address"},
	}}
}

func (r *SampleInvoiceCommand) Handle(ctx console.Context) error {
	if to := ctx.Option("email"); to != "" {
		invoice := SampleInvoice(to)
		if err := invoice.Email(to); err != nil {
			ctx.Error("sending the invoice failed: " + err.Error())
			return err
		}
		ctx.Success("sent " + invoice.Number + " to " + to + " through " + facades.Config().GetString("mail.host"))
		return nil
	}
	document, err := SampleInvoice("Example Customer").PDF()
	if err != nil {
		ctx.Error("PDF failed: " + err.Error())
		return err
	}
	output := ctx.Option("output")
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(output, document, 0o644); err != nil {
		return err
	}
	ctx.Success("wrote " + output + " (" + strconv.Itoa(len(document)) + " bytes)")
	return nil
}
