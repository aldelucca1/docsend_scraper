package mongo

import (
	"github.com/aldelucca1/docsend_scraper/model"
	"github.com/aldelucca1/docsend_scraper/store"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// DocumentCollection is the collection the holds the document objects
const DocumentCollection = "document"

func init() {
	indexes[DocumentCollection] = []mgo.Index{
		mgo.Index{Name: "idx_document_owner", Key: []string{"owner"}},
	}
}

// GetDocuments gets the set of documents for a supplied owner
func (s *Store) GetDocuments(owner string) ([]*model.Document, error) {

	// Create the query
	query := bson.M{"owner": owner}

	// Acquire a mongodb session
	session, err := s.getSession()
	if err != nil {
		return nil, s.handleError(err)
	}
	defer session.Close()

	// Query the list of documents for the supplied owner
	docs := make([]*model.Document, 0)

	db := session.DB(s.config.db)
	c := db.C(DocumentCollection)
	q := c.Find(query).Sort("-created")

	iter := q.Iter()
	for doc := new(model.Document); iter.Next(&doc); doc = new(model.Document) {
		docs = append(docs, doc)
	}
	if err = iter.Close(); err != nil {
		return nil, err
	}

	return docs, nil
}

// GetDocument gets the document with the supplied id
func (s *Store) GetDocument(id string) (*model.Document, error) {

	// Validate the supplied ID is in fact a MongoDB ObjectID
	if !bson.IsObjectIdHex(id) {
		return nil, store.ErrNotFound
	}

	// Acquire a mongodb session
	session, err := s.getSession()
	if err != nil {
		return nil, s.handleError(err)
	}
	defer session.Close()

	// Get the Document
	var doc *model.Document

	db := session.DB(s.config.db)
	c := db.C(DocumentCollection)
	err = c.FindId(bson.ObjectIdHex(id)).One(&doc)
	if err != nil {
		return nil, s.handleError(err)
	}

	return doc, nil
}

// InsertDocument inserts the supplied document
func (s *Store) InsertDocument(sourceURL string, owner string) (*model.Document, error) {

	now := makeTimestamp()

	doc := &model.Document{
		ID:        bson.NewObjectId(),
		Owner:     owner,
		SourceURL: sourceURL,
		Status:    model.StatusPending,
		StatusDetails: []model.StatusDetail{
			model.StatusDetail{
				Message: "Request Submitted",
				Created: now,
			},
		},
		Created:     now,
		LastUpdated: now,
	}

	// Acquire a mongodb session
	session, err := s.getSession()
	if err != nil {
		return nil, s.handleError(err)
	}
	defer session.Close()

	// Insert the Document
	db := session.DB(s.config.db)
	c := db.C(DocumentCollection)
	err = c.Insert(doc)
	if err != nil {
		return nil, s.handleError(err)
	}

	return doc, nil
}

// UpdateStatus updates the document's status
func (s *Store) UpdateStatus(id string, status model.Status, message string) (*model.Document, error) {

	// Validate the supplied ID is in fact a MongoDB ObjectID
	if !bson.IsObjectIdHex(id) {
		return nil, store.ErrNotFound
	}

	now := makeTimestamp()

	// Create our update document
	update := bson.M{
		"$set": bson.M{
			"status":       status,
			"last_updated": now,
		},
	}
	if message != "" {
		update["$push"] = bson.M{
			"status_details": bson.M{
				"$each": []model.StatusDetail{
					model.StatusDetail{
						Message: message,
						Created: now,
					},
				},
				"$position": 0,
			},
		}
	}

	// Acquire a mongodb session
	session, err := s.getSession()
	if err != nil {
		return nil, s.handleError(err)
	}
	defer session.Close()

	// Find and update the document
	var doc *model.Document

	db := session.DB(s.config.db)
	c := db.C(DocumentCollection)
	_, err = c.FindId(bson.ObjectIdHex(id)).Apply(mgo.Change{Update: update, ReturnNew: true}, &doc)
	if err != nil {
		return nil, s.handleError(err)
	}

	return doc, nil
}
