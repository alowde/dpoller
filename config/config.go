package config

import "github.com/pkg/errors"

//import "github.com/alowde/dpoller/url"
import "encoding/json"
import "io/ioutil"
import "net/http"

type configSkeleton struct {
	Listen    *json.RawMessage `json:"listen-config"`
	Publish   *json.RawMessage `json:"publish-config"`
	Alert     *json.RawMessage `json:"alert-config"`
	Contacts  *json.RawMessage `json:"contacts"`
	Tests     *json.RawMessage `json:"urls"`
	ConfigURL string           `json:"config-url"`
}

var Unparsed configSkeleton

func Load() error {
	if err := staticInitialise(); err != nil {
		return errors.Wrap(err, "could not initialise static config")
	}
	if err := httpInitialise(); err != nil {
		return errors.Wrap(err, "could not initialise http config")
	}
	return nil
}

func staticInitialise() error {
	raw, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return errors.Wrap(err, "could not read config file")
	}
	if err = json.Unmarshal(raw, &Unparsed); err != nil {
		return errors.Wrap(err, "could not parse config file")
	}
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

func httpInitialise() error {
	res, err := http.Get(Unparsed.ConfigURL)
	if err != nil {
		return errors.Wrap(err, "couldn't read data from Config URL")
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	if err := json.Unmarshal(body, &Unparsed); err != nil {
		return errors.Wrap(err, "couldn't parse data from Config URL")
	}
	return nil
}
