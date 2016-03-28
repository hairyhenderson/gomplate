package aws

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// DefaultEndpoint -
const DefaultEndpoint = "http://169.254.169.254"

// Ec2Meta -
type Ec2Meta struct {
	Endpoint string
	Client   *http.Client
}

// returnDefault -
func returnDefault(def []string) string {
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

func (e *Ec2Meta) retrieveMetadata(url string, key string, def ...string) string {
	if e.Client == nil {
		e.Client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	resp, err := e.Client.Get(url)
	if err != nil {
		return returnDefault(def)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		return returnDefault(def)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body from %s: %v", url, err)
	}
	value := strings.TrimSpace(string(body))

	return value
}

// Meta -
func (e *Ec2Meta) Meta(key string, def ...string) string {
	if e.Endpoint == "" {
		e.Endpoint = DefaultEndpoint
	}

	url := e.Endpoint + "/latest/meta-data/" + key
	return e.retrieveMetadata(url, key, def...)
}

// Dynamic -
func (e *Ec2Meta) Dynamic(key string, def ...string) string {
	if e.Endpoint == "" {
		e.Endpoint = DefaultEndpoint
	}

	url := e.Endpoint + "/latest/dynamic/" + key
	return e.retrieveMetadata(url, key, def...)
}

// Region -
func (e *Ec2Meta) Region() string {
	doc := e.Dynamic("instance-identity/document", `{"region":"unknown"}`)
	obj := &InstanceDocument{
		Region: "unknown",
	}
	err := json.Unmarshal([]byte(doc), &obj)
	if err != nil {
		log.Fatalf("Unable to unmarshal JSON object %s: %v", doc, err)
	}
	return obj.Region
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
