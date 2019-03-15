package funcs

import (
	"sync"

	"github.com/hairyhenderson/gomplate/aws"
	"github.com/hairyhenderson/gomplate/conv"
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
	kms      *aws.KMS
	sts      *aws.STS
	metaInit sync.Once
	infoInit sync.Once
	kmsInit  sync.Once
	stsInit  sync.Once
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

// KMSEncrypt -
func (a *Funcs) KMSEncrypt(keyID, plaintext interface{}) (string, error) {
	a.kmsInit.Do(a.initKMS)
	return a.kms.Encrypt(conv.ToString(keyID), conv.ToString(plaintext))
}

// KMSDecrypt -
func (a *Funcs) KMSDecrypt(ciphertext interface{}) (string, error) {
	a.kmsInit.Do(a.initKMS)
	return a.kms.Decrypt(conv.ToString(ciphertext))
}

// UserID - Gets the unique identifier of the calling entity. The exact value
// depends on the type of entity making the call. The values returned are those
// listed in the aws:userid column in the Principal table
// (http://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_variables.html#principaltable)
// found on the Policy Variables reference page in the IAM User Guide.
func (a *Funcs) UserID() (string, error) {
	a.stsInit.Do(a.initSTS)
	return a.sts.UserID()
}

// Account - Gets the AWS account ID number of the account that owns or
// contains the calling entity.
func (a *Funcs) Account() (string, error) {
	a.stsInit.Do(a.initSTS)
	return a.sts.Account()
}

// ARN - Gets the AWS ARN associated with the calling entity
func (a *Funcs) ARN() (string, error) {
	a.stsInit.Do(a.initSTS)
	return a.sts.Arn()
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

func (a *Funcs) initKMS() {
	if a.kms == nil {
		a.kms = aws.NewKMS(a.awsopts)
	}
}

func (a *Funcs) initSTS() {
	if a.sts == nil {
		a.sts = aws.NewSTS(a.awsopts)
	}
}
