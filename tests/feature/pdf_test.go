package feature

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/app-modules/pdf"
	"goravel/tests"
)

// PdfTestSuite renders real PDFs with headless Chrome; skipped when Chrome is not installed.
type PdfTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestPdfTestSuite(t *testing.T) {
	suite.Run(t, new(PdfTestSuite))
}

func (s *PdfTestSuite) TestTheInvoiceTemplateBecomesARealPdf() {
	if !pdf.Available() {
		s.T().Skip("Chrome is not installed")
	}
	invoice := pdf.SampleInvoice("Test Customer")
	html, err := invoice.HTML()
	s.Require().NoError(err)
	s.Contains(html, "64,025.00", "the total is in the document")
	s.Contains(html, "@page { size: A4", "page size comes from CSS")

	document, err := invoice.PDF()
	s.Require().NoError(err)
	s.True(bytes.HasPrefix(document, []byte("%PDF-")), "a PDF file")
	s.Greater(len(document), 10_000, "with real content")
	s.Equal(1, bytes.Count(document, []byte("/Type /Page\n"))+bytes.Count(document, []byte("/Type /Page/"))+bytes.Count(document, []byte("/Type /Page>")),
		"one A4 page")
}

func (s *PdfTestSuite) TestEmptyHtmlIsRefused() {
	_, err := pdf.FromHTML("")
	s.Error(err)
}

func (s *PdfTestSuite) TestAnInvoiceIsMailedWithItsPdfAttachedUnderItsNumber() {
	if !pdf.Available() {
		s.T().Skip("Chrome is not installed")
	}
	original := pdf.SendMail
	defer func() { pdf.SendMail = original }()
	var sent pdf.Outgoing
	var attached []byte
	pdf.SendMail = func(message pdf.Outgoing) error {
		sent = message
		attached, _ = os.ReadFile(message.Attachment)
		return nil
	}

	invoice := pdf.SampleInvoice("customer@example.com")
	s.Require().NoError(invoice.Email("customer@example.com"))

	s.Equal("customer@example.com", sent.To)
	s.Equal("Invoice "+invoice.Number+" from "+invoice.Seller, sent.Subject)
	s.Equal(invoice.Number+".pdf", filepath.Base(sent.Attachment))
	s.True(bytes.HasPrefix(attached, []byte("%PDF-")), "a real PDF is attached")
	for _, body := range []string{sent.HTML, sent.Text} {
		s.Contains(body, invoice.Number)
		s.Contains(body, invoice.Currency+" "+invoice.Total().String())
		s.Contains(body, "Hello customer@example.com")
	}
	_, err := os.Stat(sent.Attachment)
	s.True(os.IsNotExist(err), "the temporary PDF is removed")
}
