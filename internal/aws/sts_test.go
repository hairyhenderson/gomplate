package aws

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSTS(t *testing.T) {
	s := NewSTS()
	cid := &DummyCallerIdentifier{
		account: "acct",
		userID:  "uid",
		arn:     "arn",
	}
	s.identifier = func(_ context.Context) CallerIdentitifier {
		return cid
	}

	out, err := s.getCallerID(t.Context())
	require.NoError(t, err)
	assert.Equal(t, &sts.GetCallerIdentityOutput{
		Account: aws.String("acct"),
		Arn:     aws.String("arn"),
		UserId:  aws.String("uid"),
	}, out)

	assert.Equal(t, "acct", must(s.Account(t.Context())))
	assert.Equal(t, "arn", must(s.Arn(t.Context())))
	assert.Equal(t, "uid", must(s.UserID(t.Context())))

	s = NewSTS()
	cid = &DummyCallerIdentifier{
		account: "acct",
		userID:  "uid",
		arn:     "arn",
	}
	oldIDClient := identifierClient
	identifierClient = cid
	defer func() { identifierClient = oldIDClient }()

	out, err = s.getCallerID(t.Context())
	require.NoError(t, err)
	assert.Equal(t, &sts.GetCallerIdentityOutput{
		Account: aws.String("acct"),
		Arn:     aws.String("arn"),
		UserId:  aws.String("uid"),
	}, out)

	assert.Equal(t, "acct", must(s.Account(t.Context())))
	assert.Equal(t, "arn", must(s.Arn(t.Context())))
	assert.Equal(t, "uid", must(s.UserID(t.Context())))
}

func TestGetCallerIDErrors(t *testing.T) {
	s := NewSTS()
	cid := &DummyCallerIdentifier{
		account: "acct",
		userID:  "uid",
		arn:     "arn",
	}
	s.identifier = func(_ context.Context) CallerIdentitifier {
		return cid
	}

	out, err := s.Account(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "acct", out)

	s = NewSTS()
	cid = &DummyCallerIdentifier{
		err: errors.New("ERRORED"),
	}
	s.identifier = func(_ context.Context) CallerIdentitifier {
		return cid
	}

	_, err = s.Account(t.Context())
	require.EqualError(t, err, "ERRORED")
	_, err = s.UserID(t.Context())
	require.EqualError(t, err, "ERRORED")
	_, err = s.Arn(t.Context())
	require.EqualError(t, err, "ERRORED")
}

type DummyCallerIdentifier struct {
	err                  error
	account, arn, userID string
}

func (c *DummyCallerIdentifier) GetCallerIdentity(_ context.Context, _ *sts.GetCallerIdentityInput, _ ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
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
