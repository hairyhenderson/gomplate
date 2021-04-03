package aws

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/hairyhenderson/gomplate/v3/env"
)

// DefaultEndpoint -
var DefaultEndpoint = "http://169.254.169.254"

const (
	// the default region
	unknown = "unknown"
)

// Ec2Meta -
type Ec2Meta struct {
	Client   *http.Client
	cache    map[string]string
	Endpoint string
	options  ClientOptions
	nonAWS   bool
}

// NewEc2Meta -
func NewEc2Meta(options ClientOptions) *Ec2Meta {
	if endpoint := env.Getenv("AWS_META_ENDPOINT"); endpoint != "" {
		DefaultEndpoint = endpoint
	}

	return &Ec2Meta{cache: make(map[string]string), options: options}
}

// returnDefault -
func returnDefault(def []string) string {
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

func unreachable(err error) bool {
	if strings.Contains(err.Error(), "request canceled") ||
		strings.Contains(err.Error(), "no route to host") ||
		strings.Contains(err.Error(), "host is down") {
		return true
	}

	return false
}

// retrieve EC2 metadata, defaulting if we're not in EC2 or if there's a non-OK
// response. If there is an OK response, but we can't parse it, this errors
func (e *Ec2Meta) retrieveMetadata(url string, def ...string) (string, error) {
	if value, ok := e.cache[url]; ok {
		return value, nil
	}

	if e.nonAWS {
		return returnDefault(def), nil
	}

	if e.Client == nil {
		timeout := e.options.Timeout
		if timeout == 0 {
			timeout = 500 * time.Millisecond
		}
		e.Client = &http.Client{Timeout: timeout}
	}
	resp, err := e.Client.Get(url)
	if err != nil {
		if unreachable(err) {
			e.nonAWS = true
		}
		return returnDefault(def), nil
	}

	// nolint: errcheck
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		return returnDefault(def), nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to read response body from %s", url)
	}
	value := strings.TrimSpace(string(body))
	e.cache[url] = value

	return value, nil
}

// Meta -
func (e *Ec2Meta) Meta(key string, def ...string) (string, error) {
	if e.Endpoint == "" {
		e.Endpoint = DefaultEndpoint
	}

	url := e.Endpoint + "/latest/meta-data/" + key
	return e.retrieveMetadata(url, def...)
}

// Dynamic -
func (e *Ec2Meta) Dynamic(key string, def ...string) (string, error) {
	if e.Endpoint == "" {
		e.Endpoint = DefaultEndpoint
	}

	url := e.Endpoint + "/latest/dynamic/" + key
	return e.retrieveMetadata(url, def...)
}

// Region -
func (e *Ec2Meta) Region(def ...string) (string, error) {
	defaultRegion := returnDefault(def)
	if defaultRegion == "" {
		defaultRegion = unknown
	}

	doc, err := e.Dynamic("instance-identity/document", `{"region":"`+defaultRegion+`"}`)
	if err != nil {
		return "", err
	}
	obj := &InstanceDocument{
		Region: defaultRegion,
	}
	err = json.Unmarshal([]byte(doc), &obj)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to unmarshal JSON object %s", doc)
	}
	return obj.Region, nil
}

// InstanceDocument -
type InstanceDocument struct {
	PrivateIP        string `json:"privateIp"`
	AvailabilityZone string `json:"availabilityZone"`
	InstanceID       string `json:"InstanceId"`
	InstanceType     string `json:"InstanceType"`
	AccountID        string `json:"AccountId"`
	ImageID          string `json:"imageId"`
	Architecture     string `json:"architecture"`
	Region           string `json:"region"`
}
