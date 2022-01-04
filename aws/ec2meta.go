package aws

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hairyhenderson/gomplate/v4/env"
	"github.com/hairyhenderson/gomplate/v4/internal/deprecated"
)

const (
	// the default region
	unknown = "unknown"
)

// Ec2Meta -
type Ec2Meta struct {
	metadataCache       map[string]string
	dynamicdataCache    map[string]string
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
		metadataCache:    make(map[string]string),
		dynamicdataCache: make(map[string]string),
		ec2MetadataProvider: func() (EC2Metadata, error) {
			config := aws.NewConfig()
			config = config.WithHTTPClient(&http.Client{Timeout: options.Timeout})
			if endpoint := env.Getenv("AWS_META_ENDPOINT"); endpoint != "" {
				deprecated.WarnDeprecated(context.Background(), "Use AWS_EC2_METADATA_SERVICE_ENDPOINT instead of AWS_META_ENDPOINT")
				config = config.WithEndpoint(endpoint)
			}

			s, err := session.NewSession(config)
			if err != nil {
				return nil, err
			}

			return ec2metadata.New(s), nil
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

func (e *Ec2Meta) retrieveMetadata(key string, def ...string) (string, error) {
	if value, ok := e.metadataCache[key]; ok {
		return value, nil
	}

	if e.nonAWS {
		return returnDefault(def), nil
	}

	emd, err := e.ec2MetadataProvider()
	if err != nil {
		return "", err
	}

	value, err := emd.GetMetadata(key)
	if err != nil {
		if unreachable(err) {
			e.nonAWS = true
		}
		return returnDefault(def), nil
	}
	e.metadataCache[key] = value

	return value, nil
}

func (e *Ec2Meta) retrieveDynamicdata(key string, def ...string) (string, error) {
	if value, ok := e.dynamicdataCache[key]; ok {
		return value, nil
	}

	if e.nonAWS {
		return returnDefault(def), nil
	}

	emd, err := e.ec2MetadataProvider()
	if err != nil {
		return "", err
	}

	value, err := emd.GetDynamicData(key)
	if err != nil {
		if unreachable(err) {
			e.nonAWS = true
		}
		return returnDefault(def), nil
	}
	e.dynamicdataCache[key] = value

	return value, nil
}

// Meta -
func (e *Ec2Meta) Meta(key string, def ...string) (string, error) {
	return e.retrieveMetadata(key, def...)
}

// Dynamic -
func (e *Ec2Meta) Dynamic(key string, def ...string) (string, error) {
	return e.retrieveDynamicdata(key, def...)
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
