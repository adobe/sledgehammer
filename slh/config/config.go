/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package config

import (
	"io"

	"github.com/adobe/sledgehammer/slh/out"
	"github.com/adobe/sledgehammer/utils/db"

	"github.com/coreos/bbolt"

	"github.com/adobe/sledgehammer/utils/docker"
)

// Database is a simple struct that contains a bolt database.
// Used if only a db is required instead of the whole config.
type Database struct {
	DB *bolt.DB
}

// Docker is a simple struct that contains the docker client
// Used if only a docker client is required instead of the whole config
type Docker struct {
	Docker docker.Client
}

// Config is the main config struct for the app.
// All needed access handlers are defined here
type Config struct {
	db          *bolt.DB
	ownsDB      bool
	IO          *IO
	OutputType  string
	ConfigDir   string
	Output      *out.Output
	ExitCode    int
	Initialized bool
	Docker
}

// IO is a struct that abstracts away the In/Outputs for the app.
type IO struct {
	Out io.Writer
	Err io.Writer
	In  io.Reader
}

// OpenDatabase will open a connection to the database if not done yet and will return it
func (c *Config) OpenDatabase() (*bolt.DB, error) {
	if c.db == nil {
		database, err := db.Open(c.ConfigDir)
		if database != nil {
			c.db = database
			c.ownsDB = true
		}
		return c.db, err
	}
	return c.db, nil
}

// CloseDatabase will close the database if it is owned by the config
func (c *Config) CloseDatabase() error {
	if c.db != nil && c.ownsDB {
		err := c.db.Close()
		c.db = nil
		return err
	}
	return nil
}

// WithDatabase will create a new config with the given database and returns a new config
func (c *Config) WithDatabase(database *bolt.DB) *Config {
	return &Config{
		ownsDB:     false,
		ConfigDir:  c.ConfigDir,
		Docker:     c.Docker,
		IO:         c.IO,
		Output:     c.Output,
		OutputType: c.OutputType,
		db:         database,
	}
}

// NewOutput will create a new output container that will hold the text to render
func NewOutput(cfg *Config) *out.Output {
	o := &out.Output{
		Writer: cfg.IO.Out,
	}
	switch cfg.OutputType {
	default:
		o.RenderFunc = o.RenderTable
		o.ProgressFunc = o.TextProgressFunc
	case "text":
		o.RenderFunc = o.RenderTable
		o.ProgressFunc = o.TextProgressFunc
	case "table":
		o.RenderFunc = o.RenderTable
		o.ProgressFunc = o.NoProgressFunc
	case "json":
		o.RenderFunc = o.RenderJSON
		o.ProgressFunc = o.NoProgressFunc
	}
	return o
}
