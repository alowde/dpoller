// Package config receives and processes the root-level configuration and allocates configuration sections to other
// packages.
package config

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
		log.WithError(err).
			Warn("Somehow failed to close a Closer")
	}
}

type configSkeleton struct {
	Listen   *json.RawMessage `json:"listeners"`
	Publish  *json.RawMessage `json:"publishers"`
	Alert    *json.RawMessage `json:"alerters"`
	Contacts *json.RawMessage `json:"contacts"`
	Tests    *json.RawMessage `json:"urls"`
	Config   *configDetails   `json:"config"`
}

type configDetails struct {
	ConfigURL string `json:"url"`
	ConfigKey string `json:"key"`
}

func (c *configSkeleton) Validate() error {
	if Unparsed.Listen == nil {
		return errors.New("undefined listen block")
	} else if Unparsed.Publish == nil {
		return errors.New("undefined publish block")
	} else if Unparsed.Alert == nil {
		return errors.New("undefined alert block")
	} else if Unparsed.Contacts == nil {
		return errors.New("undefined contacts block")
	}
	return nil
}

func (c *configSkeleton) Load(r io.Reader) error {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return errors.Wrap(err, "failed to read config data from provided io.Reader")
	}
	if err := yaml.Unmarshal(buf.Bytes(), c); err != nil {
		log.WithField("config data", buf.String()).
			Debug("Failed to parse provided config data")
		return errors.Wrap(err, "failed to parse provided config data")
	}
	return nil
}

// Unparsed holds the collection of raw JSON that are subsequently parsed by other modules.
var Unparsed = configSkeleton{}

var log *logrus.Entry

// Initialise retrieves static config from local file(s) and dynamic config from an HTTP server.
func Initialise(ll logrus.Level) error {

	log = logger.New("config", ll)

	// Attempt to open and merge config from each of the provided file names
	log.Debug("Loading file configuration")
	// TODO: support runtime-defined config name
	confNames := [3]string{"config.json", "config.yaml", "config.yml"}
	for _, name := range confNames {
		file, err := os.Open(name)
		if err != nil {
			log.WithError(err).
				WithField("file", name).
				Debug("couldn't read config file")
			continue
		}
		if err := Unparsed.Load(file); err != nil {
			log.WithError(err).
				WithField("file", name).
				Warn("couldn't parse config file")
		}
	}

	// Attempt to receive and merge config from the HTTP URL, if any
	log.Debug("Loading http configuration")
	if Unparsed.Config.ConfigURL != "" {
		res, err := http.Get(Unparsed.Config.ConfigURL)
		if err != nil {
			log.WithError(err).
				WithField("url", Unparsed.Config.ConfigURL).
				Warn("couldn't read config from URL")
			return Unparsed.Validate()

		}
		defer logClose(res.Body)
		if err := Unparsed.Load(res.Body); err != nil {
			log.WithError(err).
				WithField("url", Unparsed.Config.ConfigURL).
				Warn("couldn't parse config from URL")
			return Unparsed.Validate()
		}
	}
	return Unparsed.Validate()
}
