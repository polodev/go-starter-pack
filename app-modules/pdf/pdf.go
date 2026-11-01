package pdf

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"goravel/app/facades"
)

// Options for one PDF. Page size and margins come from the CSS (@page { size: A4; margin: 18mm }), like a browser.
type Options struct {
	Timeout    time.Duration // default 30 s
	Landscape  bool
	Background bool // print background colours and images (default true through FromHTML)
}

// ChromePath is PDF_CHROME_PATH, else Chrome's usual places (chromedp finds it on the PATH and in /Applications).
func ChromePath() string {
	path, _ := facades.Config().Env("PDF_CHROME_PATH", "").(string)
	return path
}

// Available reports whether a Chrome or Chromium can be found.
func Available() bool {
	if path := ChromePath(); path != "" {
		_, err := os.Stat(path)
		return err == nil
	}
	for _, candidate := range []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/usr/bin/google-chrome", "/usr/bin/google-chrome-stable", "/usr/bin/chromium", "/usr/bin/chromium-browser",
	} {
		if _, err := os.Stat(candidate); err == nil {
			return true
		}
	}
	return false
}

// FromHTML renders HTML (with its CSS) to PDF bytes in headless Chrome.
func FromHTML(html string, options ...Options) ([]byte, error) {
	option := Options{Timeout: 30 * time.Second, Background: true}
	if len(options) > 0 {
		option = options[0]
		if option.Timeout == 0 {
			option.Timeout = 30 * time.Second
		}
	}
	if html == "" {
		return nil, errors.New("no HTML to render")
	}
	allocatorOptions := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.DisableGPU)
	if path := ChromePath(); path != "" {
		allocatorOptions = append(allocatorOptions, chromedp.ExecPath(path))
	}
	allocator, cancelAllocator := chromedp.NewExecAllocator(context.Background(), allocatorOptions...)
	defer cancelAllocator()
	browser, cancelBrowser := chromedp.NewContext(allocator)
	defer cancelBrowser()
	ctx, cancel := context.WithTimeout(browser, option.Timeout)
	defer cancel()

	// Load the HTML straight into a blank tab (no file, no web server), then print it.
	if err := chromedp.Do(ctx, chromedp.Navigate("about:blank")); err != nil {
		return nil, err
	}
	frames, err := chromedp.Call(ctx, page.GetFrameTree, cdp.Empty{})
	if err != nil {
		return nil, err
	}
	if _, err := chromedp.Call(ctx, page.SetDocumentContent, page.SetDocumentContentParams{
		FrameID: frames.FrameTree.Frame.ID, HTML: html,
	}); err != nil {
		return nil, err
	}
	pdfOptions := []chromedp.PDFOption{chromedp.PDFPreferCSSPageSize()}
	if option.Background {
		pdfOptions = append(pdfOptions, chromedp.PDFPrintBackground())
	}
	if option.Landscape {
		pdfOptions = append(pdfOptions, chromedp.PDFLandscape())
	}
	output, err := chromedp.Run(ctx, chromedp.PrintToPDF(pdfOptions...))
	if err != nil {
		return nil, err
	}
	return output, nil
}
