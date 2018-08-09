package task

import (
	"net/url"

	"github.com/aldelucca1/docsend_scraper/scraper"
	"github.com/aldelucca1/docsend_scraper/store"
)

type scrapeTask struct {
	os       store.ObjectStore
	id       string
	url      *url.URL
	email    string
	passcode string
}

// NewScrapeTask creates a new task for scraping the given URL
func NewScrapeTask(os store.ObjectStore, id string, url *url.URL, email string, passcode string) Task {
	task := &scrapeTask{
		os:       os,
		id:       id,
		url:      url,
		email:    email,
		passcode: passcode,
	}
	return task
}

func (t *scrapeTask) ID() string {
	return t.id
}

func (t *scrapeTask) Execute(status chan<- TaskStatus) error {
	s := scraper.NewScraper(t.os)
	s.StatusHandler = func(msg string) {
		status <- TaskStatus{Message: msg, Task: t}
	}
	return s.Scrape(t.url, t.email, t.passcode)
}
