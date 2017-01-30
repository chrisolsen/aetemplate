package core

import (
	"errors"
	"fmt"

	"github.com/chrisolsen/fbgraphapi"
	"golang.org/x/net/context"
)

type AuthService struct {
	URLGetter URLGetter
}

type AuthFunc func(c context.Context, creds *Credentials) (*Token, error)

// Authenticate validates that the credentials match an account; if so creates
// and links a new token to the account
// POST /v1/auth
//  {
//  	"providerName": "facebook",
//  	"providerId": "users-provider-id",
//  	"providerToken": "provided-token"
//  }
func (s *AuthService) Authenticate(c context.Context, creds *Credentials) (*Token, error) {
	var err error
	tokenStore := NewTokenStore()
	accountStore := NewAccountStore()

	switch creds.ProviderName {
	case "facebook":
		err = fbgraphapi.Authenticate(creds.ProviderToken, creds.ProviderID, s.URLGetter)
	default:
		return nil, errors.New("unknown auth provider")
	}
	if err != nil {
		return nil, fmt.Errorf("authenticate: %v", err)
	}

	accountKey, err := accountStore.GetAccountKeyByCredentials(c, creds)
	if err != nil {
		return nil, fmt.Errorf("getting account key by credentials: %v", err)
	}

	token, err := tokenStore.Create(c, accountKey)
	if err != nil {
		return nil, err
	}

	return token, nil
}
