package service

import (
	"io"

	"github.com/aldelucca1/docsend_scraper/model"
	logger "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

// Client represents a websocket client
type Client struct {
	connection *websocket.Conn
	ch         chan model.Message
	close      chan bool
}

// NewClient creates a new websocket client
func NewClient(ws *websocket.Conn) Client {
	ch := make(chan model.Message, 100)
	close := make(chan bool)

	return Client{ws, ch, close}
}

func (c *Client) listen() {
	go c.listenToWrite()
	c.listenToRead()
}

func (c *Client) listenToWrite() {
	for {
		select {
		case msg := <-c.ch:
			logger.Debugf("Send: %+v", msg)
			websocket.JSON.Send(c.connection, msg)

		case <-c.close:
			c.close <- true
			return
		}
	}
}

func (c *Client) listenToRead() {
	logger.Debug("Listening read from client")
	for {
		select {
		case <-c.close:
			c.close <- true
			return

		default:
			var msg model.Message
			err := websocket.JSON.Receive(c.connection, &msg)
			logger.Debugf("Received: %+v", msg)
			if err == io.EOF {
				c.close <- true
			} else if err != nil {
				// c.server.Err(err)
			} else if msg.Type == "PING" {
				c.ch <- model.Message{Type: "PONG", Data: nil}
			}
		}
	}
}
