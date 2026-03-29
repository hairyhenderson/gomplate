package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

var describerClient InstanceDescriber

// Ec2Info -
type Ec2Info struct {
	describer  func(context.Context) (InstanceDescriber, error)
	metaClient *Ec2Meta
	cache      map[string]any
}

// InstanceDescriber - A subset of ec2iface.EC2API that we can use to call EC2.DescribeInstances
type InstanceDescriber interface {
	DescribeInstances(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

// NewEc2Info -
func NewEc2Info() (info *Ec2Info) {
	metaClient := NewEc2Meta()
	return &Ec2Info{
		describer: func(ctx context.Context) (InstanceDescriber, error) {
			if describerClient == nil {
				describerClient = ec2.NewFromConfig(SDKConfig(ctx))
			}
			return describerClient, nil
		},
		metaClient: metaClient,
		cache:      make(map[string]any),
	}
}

// Tag -
func (e *Ec2Info) Tag(ctx context.Context, tag string, def ...string) (string, error) {
	output, err := e.describeInstance(ctx)
	if err != nil {
		return "", err
	}
	if output == nil {
		return returnDefault(def), nil
	}

	if len(output.Reservations) > 0 &&
		len(output.Reservations[0].Instances) > 0 &&
		len(output.Reservations[0].Instances[0].Tags) > 0 {
		for _, v := range output.Reservations[0].Instances[0].Tags {
			if *v.Key == tag {
				return *v.Value, nil
			}
		}
	}

	return returnDefault(def), nil
}

func (e *Ec2Info) Tags(ctx context.Context) (map[string]string, error) {
	tags := map[string]string{}

	output, err := e.describeInstance(ctx)
	if err != nil {
		return tags, err
	}
	if output == nil {
		return tags, nil
	}

	if len(output.Reservations) > 0 &&
		len(output.Reservations[0].Instances) > 0 &&
		len(output.Reservations[0].Instances[0].Tags) > 0 {
		for _, v := range output.Reservations[0].Instances[0].Tags {
			tags[*v.Key] = *v.Value
		}
	}

	return tags, nil
}

func (e *Ec2Info) describeInstance(ctx context.Context) (output *ec2.DescribeInstancesOutput, err error) {
	// cache the InstanceDescriber here
	d, err := e.describer(ctx)
	if err != nil || e.metaClient.nonAWS {
		return nil, err
	}

	if cached, ok := e.cache["DescribeInstances"]; ok {
		output = cached.(*ec2.DescribeInstancesOutput)
	} else {
		instanceID, err := e.metaClient.Meta(ctx, "instance-id")
		if err != nil {
			return nil, err
		}
		input := &ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		}

		output, err = d.DescribeInstances(ctx, input)
		if err != nil {
			// default to nil if we can't describe the instance - this could be for any reason
			return nil, nil
		}
		e.cache["DescribeInstances"] = output
	}
	return output, nil
}
