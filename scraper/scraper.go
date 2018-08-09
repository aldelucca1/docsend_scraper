package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/PuerkitoBio/goquery"
	"github.com/aldelucca1/docsend_scraper/store"
	"github.com/headzoo/surf"
	"github.com/headzoo/surf/agent"
	"github.com/headzoo/surf/browser"
)

var (
	NoopStatusHandler = func(msg string) {}
)

// Scraper is an instance of a web scraper
type Scraper struct {
	bow           *browser.Browser
	os            store.ObjectStore
	StatusHandler StatusHandler
}

// NewScraper returns a new Scraper object
func NewScraper(os store.ObjectStore) *Scraper {

	s := new(Scraper)
	s.StatusHandler = NoopStatusHandler
	s.os = os

	// Setup our Browser instance
	bow := surf.NewBrowser()
	bow.SetUserAgent(agent.Chrome())
	bow.SetAttributes(browser.AttributeMap{
		browser.SendReferer:         surf.DefaultSendReferer,
		browser.MetaRefreshHandling: false,
		browser.FollowRedirects:     true,
	})
	s.bow = bow

	return s
}

// Scrape the specified URL downloading each page image and producing a
// downloadable PDF document
func (s *Scraper) Scrape(url *url.URL, email string, passcode string) error {

	// Update the status
	s.StatusHandler("Started capturing document")

	// Open the root URL
	err := s.bow.Open(url.String())
	if err != nil {
		return err
	}
	if s.bow.StatusCode() != http.StatusOK {
		return errors.New("Failed to fetch document")
	}

	// If we have the auth form, set the supplied email and passcode
	form, err := s.bow.Form("form.new_link_auth_form")
	if err == nil {

		// Update the status
		s.StatusHandler("Authenticating with supplied email and passcode")

		// Submit the auth form
		form.Input("link_auth_form[email]", email)
		form.Input("link_auth_form[passcode]", passcode)
		err = form.Submit()
		if err != nil {
			return err
		}

		// Check if we get past after Submitting
		if _, err = s.bow.Form("form.new_link_auth_form"); err == nil {
			return errors.New("Authentication failed")
		}
	}

	// Fetch the Page data
	pages, err := s.FetchPages(s.extractImgSrc(s.bow.Dom()))
	if err != nil {
		return err
	}

	// Generate the PDF
	pdf, err := s.Generate(pages)
	if err != nil {
		return err
	}

	// Update the status
	s.StatusHandler("Writing PDF document to object storage")

	// Write the file
	pr, pw := io.Pipe()
	go func() {
		pdf.OutputAndClose(pw)
	}()

	return s.os.Write(path.Join(email, path.Base(url.Path)+".pdf"), pr)
}

// FetchPages downloads the Page information for each page container found in
// DOM
func (s *Scraper) FetchPages(urls []string) ([]*Page, error) {
	n := len(urls)
	pages := make([]*Page, n)
	for i, url := range urls {
		s.StatusHandler(fmt.Sprintf("Fetching page %d of %d", i+1, n))

		page, err := s.fetch(url, i)
		if err != nil {
			return nil, err
		}
		pages[i] = page
	}
	return pages, nil
}

func (s *Scraper) fetch(url string, index int) (*Page, error) {

	err := s.bow.Open(url)
	if err != nil {
		return nil, err
	}
	if s.bow.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("Failed to fetch page metadata for page: %d", index+1)
	}

	// Set up the pipe to write data directly into the Reader.
	pr, pw := io.Pipe()

	// Download the data through the json decoder
	go func() {
		s.bow.Download(pw)
	}()

	// Decode the read stream
	var page *Page
	err = json.NewDecoder(pr).Decode(&page)
	if err != nil {
		return nil, err
	}

	return page, nil
}

func (s *Scraper) extractImgSrc(dom *goquery.Selection) []string {
	imgs := dom.Find("div.item").Find("img.page-view").Map(func(_ int, s *goquery.Selection) string {
		if url, ok := s.Attr("data-url"); ok {
			return url
		}
		return ""
	})
	return imgs
}
