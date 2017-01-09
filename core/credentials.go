package core

import (
	"errors"
	"fmt"

	"github.com/chrisolsen/aestore"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// AuthProvider is a child to Account
type Credentials struct {
	Model

	// passed in on initial signup since looking up credentials by non-key cols
	// may result in an empty dataset
	AccountKey *datastore.Key `json:"accountKey" datastore:"-"`

	// oauth
	ProviderID   string `json:"providerId"`
	ProviderName string `json:"providerName"`

	// token is not saved
	ProviderToken string `json:"providerToken" datastore:"-"`

	// username / password
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c *Credentials) Valid() bool {
	p := len(c.ProviderID) > 0 && len(c.ProviderName) > 0 && len(c.ProviderToken) > 0
	l := len(c.Username) > 0 && len(c.Password) > 0
	return !p || !l
}

type credentialStore struct {
	aestore.Base
}

func newCredentialStore() credentialStore {
	s := credentialStore{}
	s.TableName = "credentials"
	return s
}

func (s *credentialStore) Create(c context.Context, creds *Credentials, accountKey *datastore.Key) (*datastore.Key, error) {
	if !creds.Valid() {
		return nil, errors.New("Invalid credentials")
	}

	q := datastore.NewQuery(s.TableName)
	q.KeysOnly()
	if len(creds.ProviderID) > 0 {
		q.Filter("ProviderID =", creds.ProviderID)
		q.Filter("ProviderName =", creds.ProviderName)
	} else {
		q.Filter("Username =", creds.Username)
	}
	keys, err := q.GetAll(c, nil)
	if err != nil {
		if err != datastore.ErrInvalidEntityType {
			return nil, err
		}
	}
	if len(keys) > 0 {
		return nil, errors.New("account credentials already exists")
	}

	return s.Base.Create(c, creds, accountKey)
}

func (s *credentialStore) GetAccountKeyByProvider(c context.Context, creds *Credentials) (*datastore.Key, error) {
	keys, err := datastore.NewQuery(s.TableName).
		Filter("ProviderID =", creds.ProviderID).
		Filter("ProviderName =", creds.ProviderName).
		KeysOnly().
		GetAll(c, nil)

	if err != nil {
		return nil, fmt.Errorf("finding account by auth provider: %v", err)
	}

	if len(keys) == 0 {
		return nil, errors.New("no account found matching the auth provider")
	}

	return keys[0].Parent(), nil
}
