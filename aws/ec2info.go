package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Ec2Info -
type Ec2Info struct {
	describer  InstanceDescriber
	metaClient *Ec2Meta
}

// InstanceDescriber - A subset of ec2iface.EC2API that we can use to call EC2.DescribeInstances
type InstanceDescriber interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

// var _ InstanceDescriber = (*ec2.EC2)(nil)

// NewEc2Info -
func NewEc2Info() *Ec2Info {
	metaClient := &Ec2Meta{}
	region := metaClient.Region()
	return &Ec2Info{
		describer:  ec2Client(region),
		metaClient: metaClient,
	}
}

func ec2Client(region string) (client InstanceDescriber) {
	config := aws.NewConfig()
	config = config.WithRegion(region)
	client = ec2.New(session.New(config))
	return client
}

// Tag -
func (e *Ec2Info) Tag(tag string, def ...string) string {
	instanceID := e.metaClient.Meta("instance-id")
	input := &ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{instanceID}),
	}
	output, err := e.describer.DescribeInstances(input)
	if err != nil {
		return returnDefault(def)
	}
	if output != nil && len(output.Reservations) > 0 &&
		len(output.Reservations[0].Instances) > 0 &&
		len(output.Reservations[0].Instances[0].Tags) > 0 {
		for _, v := range output.Reservations[0].Instances[0].Tags {
			if *v.Key == tag {
				return *v.Value
			}
		}
	}

	return returnDefault(def)
}
