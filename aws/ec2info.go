package aws

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var describerClient InstanceDescriber

var (
	co ClientOptions
	coInit sync.Once
	sdkSession *session.Session
	sdkSessionInit sync.Once
)

// ClientOptions -
type ClientOptions struct {
	Timeout time.Duration
}

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

// Centralised reading of AWS_TIMEOUT
// ... but cannot use in vault/auth.go as different strconv.Atoi error handling
func GetClientOptions() (ClientOptions) {
	coInit.Do(func() {
		timeout := os.Getenv("AWS_TIMEOUT")
		if timeout != "" {
			t, err := strconv.Atoi(timeout)
			if err != nil {
				log.Fatalf("Invalid AWS_TIMEOUT value '%s' - must be an integer\n", timeout)
			}

			co.Timeout = time.Duration(t) * time.Millisecond
		}
	})
	return co
}

func SDKSession() (*session.Session) {
	sdkSessionInit.Do(func() {
		options := GetClientOptions()
		timeout := options.Timeout
		if timeout == 0 {
			timeout = 500 * time.Millisecond
		}

		config := aws.NewConfig()
		config = config.WithHTTPClient(&http.Client{Timeout: timeout})

		// Waiting for https://github.com/aws/aws-sdk-go/issues/1103
		metaClient := NewEc2Meta(options)
		metaRegion := metaClient.Region()
		_, default1 := os.LookupEnv("AWS_REGION");
		_, default2 := os.LookupEnv("AWS_DEFAULT_REGION");
		if metaRegion != "unknown" && !default1 && !default2 {
			config = config.WithRegion(metaRegion)
		}

		sdkSession = session.Must(session.NewSessionWithOptions(session.Options{
			Config: *config,
			SharedConfigState: session.SharedConfigEnable,
		}))
	})
	return sdkSession
}

// NewEc2Info -
func NewEc2Info(options ClientOptions) *Ec2Info {
	metaClient := NewEc2Meta(options)
	return &Ec2Info{
		describer: func() InstanceDescriber {
			if describerClient == nil {
				describerClient = ec2.New(SDKSession())
			}
			return describerClient
		},
		metaClient: metaClient,
		cache:      make(map[string]interface{}),
	}
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
	// cache the InstanceDescriber here
	e.describer()
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
