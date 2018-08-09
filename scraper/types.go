package scraper

// StatusHandler is a handler function for status updates
type StatusHandler func(message string)

// Link represents a Link within a Page
type Link struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
	URI        string  `json:"uri"`
	TrackedURL string  `json:"trackedUrl"`
}

// Page stores details about a given Page
type Page struct {
	ImageURL       string `json:"imageUrl"`
	DirectImageURL string `json:"directImageUrl"`
	Links          []Link `json:"documentLinks"`
}
