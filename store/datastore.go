package store

import (
	"github.com/aldelucca1/docsend_scraper/model"
)

// Datastore is an interface to a backing persistent store
type Datastore interface {

	// Connect to the underlying Datastore
	Connect() error

	// Close the Datastore connection
	Close()

	// Gets the set of documents for a supplied owner
	GetDocuments(owner string) ([]*model.Document, error)

	// Gets the document with the supplied id
	GetDocument(id string) (*model.Document, error)

	// Inserts the supplied document
	InsertDocument(sourceURL string, owner string) (*model.Document, error)

	// Updates the document's status
	UpdateStatus(id string, status model.Status, message string) (*model.Document, error)
}
