package aws

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// DefaultEndpoint -
const DefaultEndpoint = "http://169.254.169.254"

// Ec2Meta -
type Ec2Meta struct {
	Endpoint string
	Client   *http.Client
}

func (e *Ec2Meta) retrieveMetadata(url string, key string, def ...string) string {
	if e.Client == nil {
		e.Client = &http.Client{}
	}
	resp, err := e.Client.Get(url)
	if err != nil {
		log.Fatalf("Failed to GET from %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		if len(def) > 0 {
			return def[0]
		}
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body from %s: %v", url, err)
	}
	value := strings.TrimSpace(string(body))

	return value
}

// Ec2meta -
func (e *Ec2Meta) Ec2meta(key string, def ...string) string {
	if e.Endpoint == "" {
		e.Endpoint = DefaultEndpoint
	}

	url := e.Endpoint + "/latest/meta-data/" + key
	return e.retrieveMetadata(url, key, def...)
}

// Ec2dynamic -
func (e *Ec2Meta) Ec2dynamic(key string, def ...string) string {
	if e.Endpoint == "" {
		e.Endpoint = DefaultEndpoint
	}

	url := e.Endpoint + "/latest/dynamic/" + key
	return e.retrieveMetadata(url, key, def...)
}
