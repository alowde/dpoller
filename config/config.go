// Package config receives and processes the root-level configuration and allocates configuration sections to other
// packages.
package config

import (
	"bytes"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/alowde/dpoller/crypto"
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

// Skeleton contains raw configuration data for use by various modules throughout the application. It explicitly
// contains opaque configuration data with the exception of the configDetails, which is for use by this package.
type Skeleton struct {
	Listen   json.RawMessage `json:"listeners"`
	Publish  json.RawMessage `json:"publishers"`
	Alert    json.RawMessage `json:"alerters"`
	Contacts json.RawMessage `json:"contacts"`
	Tests    json.RawMessage `json:"urls"`
	Config   configDetails   `json:"config"`
	log      *logrus.Entry
}

type configDetails struct {
	URL       string `json:"url"`
	Key       string `json:"key"`
	Encrypted string `json:"encrypted"`
}

// Validate performs basic sanity checking of each provided config section
func (s *Skeleton) Validate() error {
	if s.Listen == nil {
		return errors.New("undefined listen block")
	} else if s.Publish == nil {
		return errors.New("undefined publish block")
	} else if s.Alert == nil {
		return errors.New("undefined alert block")
	} else if s.Contacts == nil {
		return errors.New("undefined contacts block")
	}
	return nil
}

func (s *Skeleton) load(r io.Reader) (err error) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return errors.Wrap(err, "failed to read config data from provided io.Reader")
	}
	if err := yaml.Unmarshal(buf.Bytes(), s); err != nil {
		s.log.WithField("config data", buf.String()).
			Debug("Failed to parse provided config data")
		return errors.Wrap(err, "failed to parse provided config data")
	}
	return nil
}

func (s *Skeleton) loadFiles(filenames []string) (err error) {
	s.log.Debug("Loading file configuration")
	var ok bool
	for _, name := range filenames {
		file, err := os.Open(name)
		if err != nil {
			s.log.WithError(err).
				WithField("file", name).
				Debug("couldn't read config file")
			continue
		}
		if err := s.load(file); err != nil {
			s.log.WithError(err).
				WithField("file", name).
				Warn("couldn't parse config file")
			continue
		}
		ok = true
	}
	if !ok {
		return errors.New("Unable to parse any config files")
	}
	return nil
}

func (s *Skeleton) loadHTTP() (err error) {
	s.log.Debug("Loading http configuration")
	if s.Config.URL == "" {
		return errors.New("Invalid config URL")
	}
	res, err := http.Get(s.Config.URL)
	if err != nil {
		s.log.WithError(err).
			WithField("url", s.Config.URL).
			Warn("couldn't read config from URL")
	}
	defer logClose(res.Body)
	return s.load(res.Body)
}

func (s *Skeleton) loadEncrypted() (err error) {
	switch {
	case s.Config.Key != "" && s.Config.Encrypted == "":
		return errors.New("found encryption key but no encrypted config")
	case s.Config.Key == "" && s.Config.Encrypted != "":
		return errors.New("found encrypted config but no encryption key")
	case s.Config.Key == "" && s.Config.Encrypted == "":
		// Lack of encrypted config is not an error
		return nil
	}

	var key *[32]byte
	if key, err = crypto.Stretch(s.Config.Key); err != nil {
		return errors.Wrap(err, "could not determine key from passphrase")
	}

	plaintext, err := crypto.Decrypt64(s.Config.Encrypted, key)
	if err != nil {
		s.log.WithError(err).
			Warn("couldn't decrypt encrypted configuration data")
	}
	if err := s.load(bytes.NewReader(plaintext)); err != nil {
		s.log.WithError(err).
			Warn("couldn't parse decrypted configuration data")
	}
	return nil
}

// NewSkeleton returns a configuration skeleton with data loaded from environment variables, filenames, HTTP, etc.
// This may return a nil Skeleton and it's the caller's responsibility to call Validate() on the returned value or
// otherwise validate the data provided.
func NewSkeleton(ll logrus.Level) Skeleton {

	S := Skeleton{}

	S.log = logger.New("config", ll)

	confNames := []string{"config.json", "config.yaml", "config.yml"}

	if err := S.loadFiles(confNames); err != nil {
		S.log.WithError(err).
			WithField("filenames", confNames).
			Warn("couldn't load config from files")
	}

	if err := S.loadHTTP(); err != nil {
		S.log.WithError(err).
			WithField("url", S.Config.URL).
			Warn("couldn't load config from URL")
	}

	if err := S.loadEncrypted(); err != nil {
		S.log.WithError(err).
			Warn("error while processing encrypted config")
	}

	return S
}
