package mongo

import (
	"time"

	"github.com/aldelucca1/docsend_scraper/store"
	"github.com/globalsign/mgo"
	logger "github.com/sirupsen/logrus"
)

const (
	// MaxPageSize - The maximum number of items that can be returned in a single
	// page
	MaxPageSize = 100
	// errorCodeDuplicateKey - The MongoDB error code for duplicate key
	errorCodeDuplicateKey = 11000
)

var indexes = make(map[string][]mgo.Index)

// NewStore - Create a new MongoDB datastore
func NewStore(config *Config) *Store {
	m := new(Store)
	m.config = config
	return m
}

// Store is a Datastore backed by MongoDB
type Store struct {
	config  *Config
	session *mgo.Session
}

// Connect - Create a connection to MongoDB
func (m *Store) Connect() error {
	return m.connect()
}

// Close - Close the connection to MongoDB
func (m *Store) Close() {
	if m.session != nil {
		m.session.Close()
	}
}

func (m *Store) getSession() (*mgo.Session, error) {

	// If we failed to connect initially, retry
	if m.session == nil {
		if err := m.connect(); err != nil {
			return nil, err
		}
	}

	// Session copy is the suggested way to manage the underlying session pooling
	// https://stackoverflow.com/questions/26574594/best-practice-to-maintain-a-mgo-session
	session := m.session.Copy()
	return session, nil
}

func (m *Store) connect() error {

	logger.Infof("Connecting to mongodb at: %s", m.config.endpoints)

	// Connect
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:          m.config.endpoints,
		Database:       m.config.db,
		ReplicaSetName: m.config.replicaSet,
	})
	if err != nil {
		logger.Errorf("Failed to connect to mongodb: %s", err.Error())
		return err
	}

	// Authenticate
	if credentials := m.config.credentials; credentials != nil {
		err = session.Login(&mgo.Credential{
			Username: credentials.username,
			Password: credentials.password,
			Source:   m.config.db,
		})
		if err != nil {
			logger.Errorf("Failed to authenticate with mongodb: %s", err.Error())
			return err
		}
	}

	session.SetMode(m.config.consistency, true)

	m.session = session
	m.ensureIndexes()

	return nil
}

func (m *Store) ensureIndexes() {
	for colName, indexList := range indexes {
		c := m.session.DB(m.config.db).C(colName)
		for _, i := range indexList {
			logger.Debugf("Ensuring index: %s, on collection: %s", i.Key, colName)

			if err := c.EnsureIndex(i); err != nil {
				logger.Warnf("Failed to ensure index: %s on %s, error = %s", i.Name, i.Key, err.Error())
			}
		}
	}
}

// makeTimestamp - Generates a timestamp from the current time
func makeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

// handleError - Convert the native MongoDB error to a datastore error
func (m *Store) handleError(err error) error {
	if err == mgo.ErrNotFound {
		return store.ErrNotFound
	}
	if lerr, ok := err.(*mgo.LastError); ok && lerr.Code == errorCodeDuplicateKey {
		return store.ErrDuplicateKey
	}
	return store.ErrInternal
}
