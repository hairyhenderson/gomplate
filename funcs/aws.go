package funcs

import (
	"sync"

	"github.com/hairyhenderson/gomplate/aws"
)

var (
	af     *Funcs
	afInit sync.Once
)

// AWSNS - the aws namespace
func AWSNS() *Funcs {
	afInit.Do(func() {
		af = &Funcs{
			awsopts: aws.GetClientOptions(),
		}
	})
	return af
}

// AWSFuncs -
func AWSFuncs(f map[string]interface{}) {
	f["aws"] = AWSNS

	// global aliases - for backwards compatibility
	f["ec2meta"] = AWSNS().EC2Meta
	f["ec2dynamic"] = AWSNS().EC2Dynamic
	f["ec2tag"] = AWSNS().EC2Tag
	f["ec2region"] = AWSNS().EC2Region
}

// Funcs -
type Funcs struct {
	meta     *aws.Ec2Meta
	info     *aws.Ec2Info
	metaInit sync.Once
	infoInit sync.Once
	awsopts  aws.ClientOptions
}

// EC2Region -
func (a *Funcs) EC2Region(def ...string) (string, error) {
	a.metaInit.Do(a.initMeta)
	return a.meta.Region(def...)
}

// EC2Meta -
func (a *Funcs) EC2Meta(key string, def ...string) (string, error) {
	a.metaInit.Do(a.initMeta)
	return a.meta.Meta(key, def...)
}

// EC2Dynamic -
func (a *Funcs) EC2Dynamic(key string, def ...string) (string, error) {
	a.metaInit.Do(a.initMeta)
	return a.meta.Dynamic(key, def...)
}

// EC2Tag -
func (a *Funcs) EC2Tag(tag string, def ...string) (string, error) {
	a.infoInit.Do(a.initInfo)
	return a.info.Tag(tag, def...)
}

func (a *Funcs) initMeta() {
	if a.meta == nil {
		a.meta = aws.NewEc2Meta(a.awsopts)
	}
}

func (a *Funcs) initInfo() {
	if a.info == nil {
		a.info = aws.NewEc2Info(a.awsopts)
	}
}
