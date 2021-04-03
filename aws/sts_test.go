package aws

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/stretchr/testify/assert"
)

func TestNewSTS(t *testing.T) {
	s := NewSTS(ClientOptions{})
	cid := &DummyCallerIdentifier{
		account: "acct",
		userID:  "uid",
		arn:     "arn",
	}
	s.identifier = func() CallerIdentitifier {
		return cid
	}

	out, err := s.getCallerID()
	assert.NoError(t, err)
	assert.EqualValues(t, &sts.GetCallerIdentityOutput{
		Account: aws.String("acct"),
		Arn:     aws.String("arn"),
		UserId:  aws.String("uid"),
	}, out)

	assert.Equal(t, "acct", must(s.Account()))
	assert.Equal(t, "arn", must(s.Arn()))
	assert.Equal(t, "uid", must(s.UserID()))

	s = NewSTS(ClientOptions{})
	cid = &DummyCallerIdentifier{
		account: "acct",
		userID:  "uid",
		arn:     "arn",
	}
	oldIDClient := identifierClient
	identifierClient = cid
	defer func() { identifierClient = oldIDClient }()

	out, err = s.getCallerID()
	assert.NoError(t, err)
	assert.EqualValues(t, &sts.GetCallerIdentityOutput{
		Account: aws.String("acct"),
		Arn:     aws.String("arn"),
		UserId:  aws.String("uid"),
	}, out)

	assert.Equal(t, "acct", must(s.Account()))
	assert.Equal(t, "arn", must(s.Arn()))
	assert.Equal(t, "uid", must(s.UserID()))
}

func TestGetCallerIDErrors(t *testing.T) {
	s := NewSTS(ClientOptions{})
	cid := &DummyCallerIdentifier{
		account: "acct",
		userID:  "uid",
		arn:     "arn",
	}
	s.identifier = func() CallerIdentitifier {
		return cid
	}

	out, err := s.Account()
	assert.NoError(t, err)
	assert.Equal(t, "acct", out)

	s = NewSTS(ClientOptions{})
	cid = &DummyCallerIdentifier{
		err: errors.New("ERRORED"),
	}
	s.identifier = func() CallerIdentitifier {
		return cid
	}

	_, err = s.Account()
	assert.EqualError(t, err, "ERRORED")
	_, err = s.UserID()
	assert.EqualError(t, err, "ERRORED")
	_, err = s.Arn()
	assert.EqualError(t, err, "ERRORED")
}

type DummyCallerIdentifier struct {
	err                  error
	account, arn, userID string
}

func (c *DummyCallerIdentifier) GetCallerIdentity(*sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
	if c.err != nil {
		return nil, c.err
	}

	out := &sts.GetCallerIdentityOutput{
		Account: aws.String(c.account),
		Arn:     aws.String(c.arn),
		UserId:  aws.String(c.userID),
	}
	return out, nil
}
