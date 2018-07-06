package config

import (
	"bytes"
	"fmt"
	"github.com/Sirupsen/logrus"
	"net/http"
	"net/http/httptest"
	"testing"
)

var goodFileConfig = `
{
  "config": {
    "url": "http://localhost:9812",
    "key": "qofibQ9FY-23YQO8H3QU23GUAEFGER"
  }
}
`
var goodHTTPConfig = `
{
  "config": {
    "encrypted": "lhafelewfl"
  },
  "listeners": {
    "listener1": "foo"
  },
  "publishers":{
    "publisher1": "foo"
  },
  "alerters":{
	"alerter1": "foo"
  },
  "contacts":{
	  "alerter1": "bar"
  },
}
`

var goodDecryptedConfig = `
{
  "urls": [
    {
      "url1": "https://example.com",
      "name": "Example",
      "alert-below": 100,
      "contacts": [
        "contact1"
      ],
      "ok-statuses": [
        200
      ],
      "alert-interval": 600,
      "test-interval": 60
    }
  ]
}
`

func init() {

}

func TestLoad(t *testing.T) {
	var tables = []string{goodFileConfig, goodHTTPConfig, goodDecryptedConfig}
	for _, conf := range tables {
		b := bytes.NewBufferString(conf)
		s := new(Skeleton)
		s.logger = logrus.NewEntry(logrus.New())

		if err := s.load(b); err != nil {
			t.Errorf("Received error %v", err)
		}
	}
}

func TestLoadHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, goodHTTPConfig)
	}))
	defer ts.Close()
	s := Skeleton{
		Config: &configDetails{URL: ts.URL},
		logger: logrus.NewEntry(logrus.New()),
	}
	if err := s.loadHTTP(); err != nil {
		t.Errorf("Received error %v", err)
	}
}
