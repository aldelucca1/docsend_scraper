package model

import "github.com/globalsign/mgo/bson"

type Status int

const (
	StatusPending   Status = iota
	StatusCapturing Status = iota
	StatusComplete  Status = iota
	StatusError     Status = iota
)

type StatusDetail struct {
	Message string
	Created int64
}

type Document struct {
	ID            bson.ObjectId  `json:"id" bson:"_id"`
	Owner         string         `json:"owner"`
	SourceURL     string         `json:"source_url" bson:"source_url"`
	URL           string         `json:"url,omitempty"`
	Status        Status         `json:"status"`
	StatusDetails []StatusDetail `json:"status_details" bson:"status_details"`
	Created       int64          `json:"created"`
	LastUpdated   int64          `json:"last_updated" bson:"last_updated"`
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
