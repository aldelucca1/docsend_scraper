package service

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"

	"github.com/aldelucca1/docsend_scraper/model"
	"github.com/aldelucca1/docsend_scraper/store"
	"github.com/aldelucca1/docsend_scraper/store/fs"
	"github.com/aldelucca1/docsend_scraper/store/mongo"
	"github.com/aldelucca1/docsend_scraper/task"
	logger "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

// Service is a controller for handling inbound requests
type Service struct {
	store             store.Datastore
	os                store.ObjectStore
	dispatcher        *task.NonBlockingDispatcher
	stopStatusChannel chan chan bool
	connections       map[string]Client
}

// NewService creates a new intialized instance of a Service
func NewService() *Service {
	svc := new(Service)
	svc.store = createStore()
	svc.os = fs.NewStore(fs.NewConfig())
	svc.dispatcher = task.NewNonBlockingDispatcher(10)
	svc.connections = make(map[string]Client)
	return svc
}

// AddClientConnection adds a new client connection
func (s *Service) AddClientConnection(owner string, conn *websocket.Conn) {
	defer conn.Close()
	client := NewClient(conn)
	s.connections[owner] = client
	client.ch <- model.Message{Type: "PING", Data: nil}
	client.listen()
}

// Start starts the service.  This includes connecting to the underlying data
// store and starting our task dispatcher
func (s *Service) Start() error {

	err := s.store.Connect()
	if err != nil {
		return err
	}

	s.stopStatusChannel = make(chan chan bool, 1)

	// Start our Dispatcher
	s.dispatcher.Start()
	go s.dispatcherStatusHandler()
	return nil
}

// Stop stops the Dispatcher, waiting for any in flight work to complete and
// closes the connection to the underlying datastore
func (s *Service) Stop() {

	if s.dispatcher != nil {
		s.dispatcher.Stop()
	}

	stoppedChan := make(chan bool, 1)
	s.stopStatusChannel <- stoppedChan
	<-stoppedChan

	s.store.Close()
}

func createStore() store.Datastore {
	config := mongo.NewConfig()
	return mongo.NewStore(config)
}

// dispatcherStatusHandler - A go routine reponsible for listening for events
// from our Dispatcher
//
// When tasks complete they will be pulled off the Dispatcher's Complete channel
// and passed to the handleTaskComplete function
//
// In the case of task failures, the failure information will be pulled off the
// Dipatcher's Error channel and passed to the handleTaskError function
//
// To stop this go routine, pass a stopped chan to the stopStatusChannel. When
// this routine completes it will notify the passed stopped channel
//
func (s *Service) dispatcherStatusHandler() {

	var stoppedChan chan bool
	var stopped bool

	for stopped != true {
		select {

		// Listen for task complete
		case task, more := <-s.dispatcher.Complete():
			if more {
				s.handleTaskComplete(task)
			}

		case status, more := <-s.dispatcher.Status():
			if more {
				s.handleTaskStatus(status)
			}

		// An error occurred, handle it accordingly
		case taskerror, more := <-s.dispatcher.Error():
			if more {
				s.handleTaskError(taskerror)
			}

		case stoppedChan = <-s.stopStatusChannel:
			stopped = true
			break
		}
	}
	stoppedChan <- true
}

func (s *Service) handleTaskStatus(status task.TaskStatus) {

	logger.Infof("Task %s has updated its status: %s", status.Task.ID(), status.Message)

	doc, err := s.store.UpdateStatus(status.Task.ID(), model.StatusCapturing, status.Message)
	if err != nil {
		logger.Errorf("Failed to store document status update: %s", err.Error())
		return
	}
	s.pushDocument(doc)
}

func (s *Service) handleTaskComplete(task task.Task) {

	logger.Infof("Task %s completed successfully", task.ID())

	doc, err := s.store.UpdateStatus(task.ID(), model.StatusComplete, "Completed successfully")
	if err != nil {
		logger.Errorf("Failed to store document status update: %s", err.Error())
		return
	}
	s.pushDocument(doc)
}

func (s *Service) handleTaskError(taskerror task.Failure) {

	logger.Infof("Task %s failed with error: %s", taskerror.Task.ID(), taskerror.Error.Error())

	doc, err := s.store.UpdateStatus(taskerror.Task.ID(), model.StatusError, fmt.Sprintf("Failed with error: %s", taskerror.Error.Error()))
	if err != nil {
		logger.Errorf("Failed to store document status update: %s", err.Error())
		return
	}
	s.pushDocument(doc)
}

func (s *Service) pushDocument(doc *model.Document) {
	if conn, ok := s.connections[doc.Owner]; ok {
		conn.ch <- model.Message{Type: "UPDATE", Data: doc}
	}
}

// ListDocuments lists the set of document metadata for the given user
func (s *Service) ListDocuments(user string) ([]*model.Document, error) {
	return s.store.GetDocuments(user)
}

// GetDocument gets the document metadata
func (s *Service) GetDocument(id string) (*model.Document, error) {
	return s.store.GetDocument(id)
}

// GenerateDocument generates the PDF from the specified source url
func (s *Service) GenerateDocument(urlStr string, email string, passcode string) (*model.Document, error) {

	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	if url.Scheme != "https" && url.Host != "docsend.com" {
		return nil, errors.New("Invalid URL")
	}

	doc, err := s.store.InsertDocument(url.String(), email)
	if err != nil {
		return nil, err
	}

	scrape := task.NewScrapeTask(s.os, doc.ID.Hex(), url, email, passcode)
	s.dispatcher.Dispatch(scrape)

	return doc, nil
}

// DownloadDocument reads the document from the object store and sends it to the
// user
func (s *Service) DownloadDocument(id string) (io.Reader, error) {

	doc, err := s.store.GetDocument(id)
	if err != nil {
		return nil, err
	}

	src := path.Join(doc.Owner, path.Base(doc.SourceURL)+".pdf")

	logger.Infof("Download document at: %s", src)

	return s.os.Read(src)
}
