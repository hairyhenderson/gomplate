package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Ec2Info -
type Ec2Info struct {
	describer  func() InstanceDescriber
	metaClient *Ec2Meta
	cache      map[string]interface{}
}

// InstanceDescriber - A subset of ec2iface.EC2API that we can use to call EC2.DescribeInstances
type InstanceDescriber interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

// NewEc2Info -
func NewEc2Info() *Ec2Info {
	metaClient := NewEc2Meta()
	return &Ec2Info{
		describer: func() InstanceDescriber {
			region := metaClient.Region()
			return ec2Client(region)
		},
		metaClient: metaClient,
		cache:      make(map[string]interface{}),
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
	output := e.describeInstance()
	if output == nil {
		return returnDefault(def)
	}

	if len(output.Reservations) > 0 &&
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

func (e *Ec2Info) describeInstance() (output *ec2.DescribeInstancesOutput) {
	if e.metaClient.nonAWS {
		return nil
	}

	if cached, ok := e.cache["DescribeInstances"]; ok {
		output = cached.(*ec2.DescribeInstancesOutput)
	} else {
		instanceID := e.metaClient.Meta("instance-id")

		input := &ec2.DescribeInstancesInput{
			InstanceIds: aws.StringSlice([]string{instanceID}),
		}

		var err error
		output, err = e.describer().DescribeInstances(input)
		if err != nil {
			return nil
		}
		e.cache["DescribeInstances"] = output
	}
	return
}
