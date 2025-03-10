package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// MockEC2Meta -
func MockEC2Meta(data map[string]string, dynamicData map[string]string, region string) *Ec2Meta {
	return &Ec2Meta{
		metadataCache:    map[string]string{},
		dynamicdataCache: map[string]string{},
		ec2MetadataProvider: func() (EC2Metadata, error) {
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
		describer:  func() (InstanceDescriber, error) { return DummyInstanceDescriber{}, nil },
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
	tags []*ec2.Tag
}

// DescribeInstances -
func (d DummyInstanceDescriber) DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	output := &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{
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

func (d DummEC2MetadataProvider) GetMetadata(p string) (string, error) {
	v, ok := d.data[p]
	if !ok {
		return "", fmt.Errorf("cannot find %v", p)
	}
	return v, nil
}

func (d DummEC2MetadataProvider) GetDynamicData(p string) (string, error) {
	v, ok := d.dynamicData[p]
	if !ok {
		return "", fmt.Errorf("cannot find %v", p)
	}
	return v, nil
}

func (d DummEC2MetadataProvider) Region() (string, error) {
	return d.region, nil
}
