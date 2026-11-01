package feature

import (
	"bytes"
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
