package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// STS -
type STS struct {
	identifier func(context.Context) CallerIdentitifier
	cache      map[string]any
}

var identifierClient CallerIdentitifier

// CallerIdentitifier - an interface to wrap GetCallerIdentity
type CallerIdentitifier interface {
	GetCallerIdentity(context.Context, *sts.GetCallerIdentityInput, ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

// NewSTS -
func NewSTS() *STS {
	return &STS{
		identifier: func(ctx context.Context) CallerIdentitifier {
			if identifierClient == nil {
				identifierClient = sts.NewFromConfig(SDKConfig(ctx))
			}
			return identifierClient
		},
		cache: make(map[string]any),
	}
}

func (s *STS) getCallerID(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
	i := s.identifier(ctx)
	if val, ok := s.cache["GetCallerIdentity"]; ok {
		if c, ok := val.(*sts.GetCallerIdentityOutput); ok {
			return c, nil
		}
	}
	in := &sts.GetCallerIdentityInput{}
	out, err := i.GetCallerIdentity(ctx, in)
	if err != nil {
		return nil, err
	}
	s.cache["GetCallerIdentity"] = out
	return out, nil
}

// UserID -
func (s *STS) UserID(ctx context.Context) (string, error) {
	cid, err := s.getCallerID(ctx)
	if err != nil {
		return "", err
	}
	return *cid.UserId, nil
}

// Account -
func (s *STS) Account(ctx context.Context) (string, error) {
	cid, err := s.getCallerID(ctx)
	if err != nil {
		return "", err
	}
	return *cid.Account, nil
}

// Arn -
func (s *STS) Arn(ctx context.Context) (string, error) {
	cid, err := s.getCallerID(ctx)
	if err != nil {
		return "", err
	}
	return *cid.Arn, nil
}
