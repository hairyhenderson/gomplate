package aws

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

var describerClient InstanceDescriber

var (
	co             ClientOptions
	coInit         sync.Once
	sdkSession     *session.Session
	sdkSessionInit sync.Once
)

const (
	unknown = "unknown"
)

// ClientOptions -
type ClientOptions struct {
	Timeout time.Duration
}

// Ec2Info -
type Ec2Info struct {
	describer  func() (InstanceDescriber, error)
	metaClient *Ec2Meta
	cache      map[string]interface{}
}

// InstanceDescriber - A subset of ec2iface.EC2API that we can use to call EC2.DescribeInstances
type InstanceDescriber interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

// GetClientOptions - Centralised reading of AWS_TIMEOUT
// ... but cannot use in vault/auth.go as different strconv.Atoi error handling
func GetClientOptions() ClientOptions {
	coInit.Do(func() {
		timeout := os.Getenv("AWS_TIMEOUT")
		if timeout != "" {
			t, err := strconv.Atoi(timeout)
			if err != nil {
				panic(errors.Wrapf(err, "Invalid AWS_TIMEOUT value '%s' - must be an integer\n", timeout))
			}

			co.Timeout = time.Duration(t) * time.Millisecond
		}
	})
	return co
}

// SDKSession -
func SDKSession(region ...string) *session.Session {
	sdkSessionInit.Do(func() {
		options := GetClientOptions()
		timeout := options.Timeout
		if timeout == 0 {
			timeout = 500 * time.Millisecond
		}

		config := aws.NewConfig()
		config = config.WithHTTPClient(&http.Client{Timeout: timeout})

		metaRegion := unknown
		if len(region) > 0 {
			metaRegion = region[0]
		}
		// Waiting for https://github.com/aws/aws-sdk-go/issues/1103
		_, default1 := os.LookupEnv("AWS_REGION")
		_, default2 := os.LookupEnv("AWS_DEFAULT_REGION")
		if metaRegion != unknown && !default1 && !default2 {
			config = config.WithRegion(metaRegion)
		}

		sdkSession = session.Must(session.NewSessionWithOptions(session.Options{
			Config:            *config,
			SharedConfigState: session.SharedConfigEnable,
		}))
	})
	return sdkSession
}

// NewEc2Info -
func NewEc2Info(options ClientOptions) (info *Ec2Info) {
	metaClient := NewEc2Meta(options)
	return &Ec2Info{
		describer: func() (InstanceDescriber, error) {
			if describerClient == nil {
				metaRegion, err := metaClient.Region()
				if err != nil {
					return nil, errors.Wrap(err, "failed to determine EC2 region")
				}
				session := SDKSession(metaRegion)
				describerClient = ec2.New(session)
			}
			return describerClient, nil
		},
		metaClient: metaClient,
		cache:      make(map[string]interface{}),
	}
}

// Tag -
func (e *Ec2Info) Tag(tag string, def ...string) (string, error) {
	output, err := e.describeInstance()
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

func (e *Ec2Info) describeInstance() (output *ec2.DescribeInstancesOutput, err error) {
	// cache the InstanceDescriber here
	d, err := e.describer()
	if err != nil || e.metaClient.nonAWS {
		return nil, err
	}

	if cached, ok := e.cache["DescribeInstances"]; ok {
		output = cached.(*ec2.DescribeInstancesOutput)
	} else {
		instanceID, err := e.metaClient.Meta("instance-id")
		if err != nil {
			return nil, err
		}
		input := &ec2.DescribeInstancesInput{
			InstanceIds: aws.StringSlice([]string{instanceID}),
		}

		output, err = d.DescribeInstances(input)
		if err != nil {
			// default to nil if we can't describe the instance - this could be for any reason
			return nil, nil
		}
		e.cache["DescribeInstances"] = output
	}
	return output, nil
}
