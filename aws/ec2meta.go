package aws

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hairyhenderson/gomplate/v3/env"
)

const (
	// the default region
	unknown = "unknown"
)

type dataType int

const (
	metaData dataType = iota
	dynamicData
)

var ec2metadataClient EC2Metadata

// Ec2Meta -
type Ec2Meta struct {
	cache               map[string]string
	ec2MetadataProvider func() (EC2Metadata, error)
	nonAWS              bool
}

type EC2Metadata interface {
	GetMetadata(p string) (string, error)
	GetDynamicData(p string) (string, error)
	Region() (string, error)
}

// NewEc2Meta -
func NewEc2Meta(options ClientOptions) *Ec2Meta {
	return &Ec2Meta{
		cache: make(map[string]string),
		ec2MetadataProvider: func() (EC2Metadata, error) {
			if ec2metadataClient == nil {
				config := aws.NewConfig()
				config = config.WithHTTPClient(&http.Client{Timeout: options.Timeout})
				if endpoint := env.Getenv("AWS_META_ENDPOINT"); endpoint != "" {
					config = config.WithEndpoint(endpoint)
				}

				s, err := session.NewSession(config)
				if err != nil {
					return nil, err
				}
				ec2metadataClient = ec2metadata.New(s)
			}
			return ec2metadataClient, nil
		},
	}
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

func (e *Ec2Meta) retrieveData(dtype dataType, key string, def ...string) (string, error) {
	if value, ok := e.cache[key]; ok {
		return value, nil
	}

	if e.nonAWS {
		return returnDefault(def), nil
	}

	emd, err := e.ec2MetadataProvider()
	if err != nil {
		return "", err
	}

	var value string
	switch dtype {
	case metaData:
		value, err = emd.GetMetadata(key)
	case dynamicData:
		value, err = emd.GetDynamicData(key)
	default:
		return "", fmt.Errorf("unknown type: %v", dtype)
	}

	if err != nil {
		if unreachable(err) {
			e.nonAWS = true
		}

		return returnDefault(def), nil
	}
	e.cache[key] = value

	return value, nil
}

// Meta -
func (e *Ec2Meta) Meta(key string, def ...string) (string, error) {
	return e.retrieveData(metaData, key, def...)
}

// Dynamic -
func (e *Ec2Meta) Dynamic(key string, def ...string) (string, error) {
	return e.retrieveData(dynamicData, key, def...)
}

// Region -
func (e *Ec2Meta) Region(def ...string) (string, error) {
	defaultRegion := returnDefault(def)
	if defaultRegion == "" {
		defaultRegion = unknown
	}

	if e.nonAWS {
		return defaultRegion, nil
	}

	emd, err := e.ec2MetadataProvider()
	if err != nil {
		return "", err
	}

	region, err := emd.Region()
	if err != nil || region == "" {
		return defaultRegion, nil
	}

	return region, nil
}
