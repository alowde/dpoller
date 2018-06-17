// Package config receives and processes the root-level configuration and allocates configuration sections to other
// packages.
package main

import (
	"bytes"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/logger"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
)

func logClose(c io.Closer) {
	if err := c.Close(); err != nil {
		logrus.WithError(err).
			Warn("Somehow failed to close a Closer")
	}
}

// Skeleton is the skeleton of a configuration. Only exists to provide first level of config unmarshal.
type Skeleton struct {
	Listen   *json.RawMessage `json:"listeners"`
	Publish  *json.RawMessage `json:"publishers"`
	Alert    *json.RawMessage `json:"alerters"`
	Contacts *json.RawMessage `json:"contacts"`
	Tests    *json.RawMessage `json:"urls"`
	Config   *configDetails   `json:"config"`
	logger   *logrus.Entry
}

type configDetails struct {
	ConfigURL string `json:"url"`
	ConfigKey string `json:"key"`
}

func (c *Skeleton) validate() error {
	if c.Listen == nil {
		return errors.New("undefined listen block")
	} else if c.Publish == nil {
		return errors.New("undefined publish block")
	} else if c.Alert == nil {
		return errors.New("undefined alert block")
	} else if c.Contacts == nil {
		return errors.New("undefined contacts block")
	}
	return nil
}

func (c *Skeleton) load(r io.Reader) error {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return errors.Wrap(err, "failed to read config data from provided io.Reader")
	}
	if err := yaml.Unmarshal(buf.Bytes(), c); err != nil {
		c.logger.WithField("config data", buf.String()).
			Debug("Failed to parse provided config data")
		return errors.Wrap(err, "failed to parse provided config data")
	}
	return nil
}

// NewSkeleton returns a new Skeleton that includes all available configuration, and an error if there's insufficient
// valid configuration available
func NewSkeleton(ll logrus.Level) (c *Skeleton, err error) {

	c = new(Skeleton)
	c.logger = logger.New("config", ll)

	// Attempt to open and merge config from each of the provided file names
	c.logger.Debug("Loading file configuration")
	// TODO: support runtime-defined config name
	confNames := [3]string{"config.json", "config.yaml", "config.yml"}
	for _, name := range confNames {
		file, err := os.Open(name)
		if err != nil {
			c.logger.WithError(err).
				WithField("file", name).
				Debug("couldn't read config file")
			continue
		}
		if err := c.load(file); err != nil {
			c.logger.WithError(err).
				WithField("file", name).
				Warn("couldn't parse config file")
		}
	}

	// Attempt to receive and merge config from the HTTP URL, if any
	c.logger.Debug("Loading http configuration")
	if c.Config.ConfigURL != "" {
		res, err := http.Get(c.Config.ConfigURL)
		if err != nil {
			c.logger.WithError(err).
				WithField("url", c.Config.ConfigURL).
				Warn("couldn't read config from URL")
			return c, c.validate()

		}
		defer logClose(res.Body)
		if err := c.load(res.Body); err != nil {
			c.logger.WithError(err).
				WithField("url", c.Config.ConfigURL).
				Warn("couldn't parse config from URL")
			return c, c.validate()
		}
	}
	return c, c.validate()
}
