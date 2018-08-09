package scraper

import (
	"fmt"
	"net/http"

	"github.com/jung-kurt/gofpdf"
	logger "github.com/sirupsen/logrus"
)

// Generate a PDF with the given set of Pages
func (s *Scraper) Generate(pages []*Page) (*gofpdf.Fpdf, error) {

	// Update the status
	s.StatusHandler("Generating the PDF document")

	// Create our PDF container
	pdf := gofpdf.New("L", "pt", "A4", "")
	pdf.SetMargins(0, 0, 0)

	// Add each page
	for i, page := range pages {
		logger.Debugf("Page %d = %+v", i, page)
		err := s.addPage(pdf, page, i)
		if err != nil {
			return nil, err
		}
	}
	return pdf, nil
}

func (s *Scraper) addPage(pdf *gofpdf.Fpdf, page *Page, index int) error {

	width, height := pdf.GetPageSize()

	// Attempt to download the image
	rsp, err := http.Get(page.ImageURL)
	if err != nil {
		return err
	}
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to fetch image for page: %d", index+1)
	}
	defer rsp.Body.Close()

	// Add the Page to the PDF
	pdf.AddPage()

	// Add the image
	contentType := pdf.ImageTypeFromMime(rsp.Header["Content-Type"][0])
	pdf.RegisterImageReader(page.ImageURL, contentType, rsp.Body)
	pdf.Image(page.ImageURL, 0, 0, width, height, false, contentType, 0, "")

	// Add each Link
	for _, link := range page.Links {
		pdf.LinkString(width*link.X, height*link.Y, width*link.Width, height*link.Height, link.URI)
	}
	return nil
}
