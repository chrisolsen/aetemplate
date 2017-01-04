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
func (s *AuthService) Authenticate(c context.Context, creds *Credentials) (*Token, error) {
	var err error
	tokenSvc := TokenService{}
	accountSvc := AccountService{}

	switch creds.ProviderName {
	case "facebook":
		err = fbgraphapi.Authenticate(creds.ProviderToken, creds.ProviderID, s.URLGetter)
	default:
		return nil, errors.New("unknown auth provider")
	}
	if err != nil {
		return nil, fmt.Errorf("authenticate: %v", err)
	}

	accountKey, err := accountSvc.GetAccountKeyByCredentials(c, creds)
	if err != nil {
		return nil, fmt.Errorf("getting account key by credentials: %v", err)
	}

	token := tokenSvc.NewToken()
	_, err = tokenSvc.AddToken(c, accountKey, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}
