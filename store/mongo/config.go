package mongo

import (
	"os"

	"github.com/globalsign/mgo"
)

// Credentials - The username and password for the defined mongodb instance
type Credentials struct {
	username string
	password string
}

// Config - The configuration information for connecting to a mongodb
// instance
type Config struct {
	endpoints   []string
	replicaSet  string
	credentials *Credentials
	consistency mgo.Mode
	db          string
}

// NewConfig - Creates a new Config with the default values
func NewConfig() *Config {
	return &Config{
		endpoints:   []string{os.Getenv("MONGO_HOST")},
		replicaSet:  "",
		db:          "docsend",
		consistency: mgo.PrimaryPreferred,
		credentials: &Credentials{username: "docsend", password: os.Getenv("MONGO_PWD")},
	}
}

// WithEndpoints - Set the set of endpoints to connect to
func (c *Config) WithEndpoints(endpoints []string) *Config {
	c.endpoints = endpoints
	return c
}

// WithReplicaSet - Set the replica set to connect to
func (c *Config) WithReplicaSet(replicaSet string) *Config {
	c.replicaSet = replicaSet
	return c
}

// WithCredentials - Set the credentials to use when connecting
func (c *Config) WithCredentials(username string, password string) *Config {
	c.credentials = &Credentials{username: username, password: password}
	return c
}

// WithDatabase - Set the name of the database to connect to
func (c *Config) WithDatabase(db string) *Config {
	c.db = db
	return c
}

// WithConsistencyLevel - Set the consistency level to use when reading/writing
func (c *Config) WithConsistencyLevel(mode mgo.Mode) *Config {
	c.consistency = mode
	return c
}
