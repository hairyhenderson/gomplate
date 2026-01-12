package aws

import (
	"context"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
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
	ec2MetadataProvider func(context.Context) (EC2Metadata, error)
	nonAWS              bool
}

type EC2Metadata interface {
	GetMetadata(context.Context, *imds.GetMetadataInput, ...func(*imds.Options)) (*imds.GetMetadataOutput, error)
	GetDynamicData(context.Context, *imds.GetDynamicDataInput, ...func(*imds.Options)) (*imds.GetDynamicDataOutput, error)
	GetRegion(context.Context, *imds.GetRegionInput, ...func(*imds.Options)) (*imds.GetRegionOutput, error)
}

// NewEc2Meta -
func NewEc2Meta() *Ec2Meta {
	return &Ec2Meta{
		metadataCache:    make(map[string]string),
		dynamicdataCache: make(map[string]string),
		ec2MetadataProvider: func(ctx context.Context) (EC2Metadata, error) {
			client := imds.NewFromConfig(SDKConfig(ctx), func(o *imds.Options) {
				if endpoint := env.Getenv("AWS_META_ENDPOINT"); endpoint != "" {
					deprecated.WarnDeprecated(ctx, "Use AWS_EC2_METADATA_SERVICE_ENDPOINT instead of AWS_META_ENDPOINT")
					o.Endpoint = endpoint
				}
			})

			return client, nil
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

func (e *Ec2Meta) retrieveMetadata(ctx context.Context, key string, def ...string) (string, error) {
	if value, ok := e.metadataCache[key]; ok {
		return value, nil
	}

	if e.nonAWS {
		return returnDefault(def), nil
	}

	emd, err := e.ec2MetadataProvider(ctx)
	if err != nil {
		return "", err
	}

	output, err := emd.GetMetadata(ctx, &imds.GetMetadataInput{Path: key})
	if err != nil {
		if unreachable(err) {
			e.nonAWS = true
		}
		return returnDefault(def), nil
	}
	defer output.Content.Close()

	result, err := io.ReadAll(output.Content)
	if err != nil {
		return "", err
	}

	value := string(result)

	e.metadataCache[key] = value

	return value, nil
}

func (e *Ec2Meta) retrieveDynamicdata(ctx context.Context, key string, def ...string) (string, error) {
	if value, ok := e.dynamicdataCache[key]; ok {
		return value, nil
	}

	if e.nonAWS {
		return returnDefault(def), nil
	}

	emd, err := e.ec2MetadataProvider(ctx)
	if err != nil {
		return "", err
	}

	output, err := emd.GetDynamicData(ctx, &imds.GetDynamicDataInput{Path: key})
	if err != nil {
		if unreachable(err) {
			e.nonAWS = true
		}
		return returnDefault(def), nil
	}
	defer output.Content.Close()

	result, err := io.ReadAll(output.Content)
	if err != nil {
		return "", err
	}

	value := string(result)

	e.dynamicdataCache[key] = value

	return value, nil
}

// Meta -
func (e *Ec2Meta) Meta(ctx context.Context, key string, def ...string) (string, error) {
	return e.retrieveMetadata(ctx, key, def...)
}

// Dynamic -
func (e *Ec2Meta) Dynamic(ctx context.Context, key string, def ...string) (string, error) {
	return e.retrieveDynamicdata(ctx, key, def...)
}

// Region -
func (e *Ec2Meta) Region(ctx context.Context, def ...string) (string, error) {
	defaultRegion := returnDefault(def)
	if defaultRegion == "" {
		defaultRegion = unknown
	}

	if e.nonAWS {
		return defaultRegion, nil
	}

	emd, err := e.ec2MetadataProvider(ctx)
	if err != nil {
		return "", err
	}

	output, err := emd.GetRegion(ctx, &imds.GetRegionInput{})
	if err != nil || output.Region == "" {
		return defaultRegion, nil
	}

	return output.Region, nil
}
