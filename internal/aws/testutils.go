package aws

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// MockEC2Meta -
func MockEC2Meta(data map[string]string, dynamicData map[string]string, region string) *Ec2Meta {
	return &Ec2Meta{
		metadataCache:    map[string]string{},
		dynamicdataCache: map[string]string{},
		ec2MetadataProvider: func(_ context.Context) (EC2Metadata, error) {
			return &DummEC2MetadataProvider{
				data:        data,
				dynamicData: dynamicData,
				region:      region,
			}, nil
		},
	}
}

// NewDummyEc2Info -
func NewDummyEc2Info(metaClient *Ec2Meta) *Ec2Info {
	i := &Ec2Info{
		metaClient: metaClient,
		describer:  func(_ context.Context) (InstanceDescriber, error) { return DummyInstanceDescriber{}, nil },
		cache:      map[string]any{},
	}
	return i
}

// NewDummyEc2Meta -
func NewDummyEc2Meta() *Ec2Meta {
	return &Ec2Meta{
		nonAWS: true,
	}
}

// DummyInstanceDescriber - test doubles
type DummyInstanceDescriber struct {
	tags []types.Tag
}

// DescribeInstances -
func (d DummyInstanceDescriber) DescribeInstances(_ context.Context, _ *ec2.DescribeInstancesInput, _ ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	output := &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						Tags: d.tags,
					},
				},
			},
		},
	}
	return output, nil
}

type DummEC2MetadataProvider struct {
	data        map[string]string
	dynamicData map[string]string
	region      string
}

func (d DummEC2MetadataProvider) GetMetadata(_ context.Context, params *imds.GetMetadataInput, _ ...func(*imds.Options)) (*imds.GetMetadataOutput, error) {
	v, ok := d.data[params.Path]
	if !ok {
		return nil, fmt.Errorf("cannot find %v", params.Path)
	}
	return &imds.GetMetadataOutput{Content: io.NopCloser(strings.NewReader(v))}, nil
}

func (d DummEC2MetadataProvider) GetDynamicData(_ context.Context, params *imds.GetDynamicDataInput, _ ...func(*imds.Options)) (*imds.GetDynamicDataOutput, error) {
	v, ok := d.dynamicData[params.Path]
	if !ok {
		return nil, fmt.Errorf("cannot find %v", params.Path)
	}
	return &imds.GetDynamicDataOutput{Content: io.NopCloser(strings.NewReader(v))}, nil
}

func (d DummEC2MetadataProvider) GetRegion(_ context.Context, _ *imds.GetRegionInput, _ ...func(*imds.Options)) (*imds.GetRegionOutput, error) {
	return &imds.GetRegionOutput{Region: d.region}, nil
}
