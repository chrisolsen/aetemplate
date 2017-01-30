package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/chrisolsen/ae/model"
	"github.com/chrisolsen/ae/store"
	"github.com/chrisolsen/async"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

const accountsTable string = "accounts"

// AccountPayload contains the account and related data
type AccountPayload struct {
	Account

	// other attributes and relations
}

// Account model
type Account struct {
	model.Base

	FirstName string `json:"firstName" datastore:",noindex"`
	LastName  string `json:"lastName" datastore:",noindex"`
	Gender    string `json:"gender" datastore:",noindex"`
	Locale    string `json:"locale" datastore:",noindex"`
	Location  string `json:"location" datastore:",noindex"`
	Name      string `json:"name" datastore:",noindex"`
	Timezone  int    `json:"timezone" datastore:",noindex"`
	Email     string `json:"email"`

	Photo Attachment `json:"photo"`
}

type AccountStore struct {
	store.Base
}

func NewAccountStore() AccountStore {
	s := AccountStore{}
	s.TableName = "accounts"
	return s
}

// Create creates a new account and creates its default subscriptions
func (s *AccountStore) Create(c context.Context, creds *Credentials, account *Account) (*datastore.Key, error) {
	var err error
	var accountKey *datastore.Key
	var cStore = NewCredentialStore()
	err = datastore.RunInTransaction(c, func(tc context.Context) error {
		accountKey, err = s.Base.Create(tc, account, nil)
		if err != nil {
			return fmt.Errorf("failed to create account: %v", err)
		}

		_, err = cStore.Create(tc, creds, accountKey)
		if err != nil {
			return fmt.Errorf("failed to create credentials: %v", err)
		}

		return nil
	}, &datastore.TransactionOptions{XG: true})

	return accountKey, nil
}

// GetAccountKeyByCredentials fetches the account matching the auth provider credentials
func (s *AccountStore) GetAccountKeyByCredentials(c context.Context, creds *Credentials) (*datastore.Key, error) {
	var err error
	cstore := NewCredentialStore()
	// on initial signup the account key will exist within the credentials
	if creds.AccountKey != nil {
		var accountCreds []*Credentials
		_, err = cstore.GetByParent(c, creds.AccountKey, &accountCreds)
		if err != nil {
			return nil, fmt.Errorf("failed to find credentials by parent account: %v", err)
		}
		// validate credentials
		for _, ac := range accountCreds {
			if ac.ProviderID == creds.ProviderID && ac.ProviderName == creds.ProviderName {
				return creds.AccountKey, nil
			}
		}
		return nil, errors.New("no matching credentials found for account")
	}

	// by provider
	if len(creds.ProviderID) > 0 {
		return cstore.GetAccountKeyByProvider(c, creds)
	}

	// by username
	var userNameCreds []*Credentials
	_, err = cstore.GetByUsername(c, creds.Username, &userNameCreds)
	if err != nil {
		return nil, err
	}

	if len(userNameCreds) != 1 {
		return nil, errors.New("unable to find unique credentials")
	}

	crypt := Crypt{}
	err = crypt.Validate(userNameCreds[0].Password, creds.Password)
	if err != nil {
		return nil, err
	}
	return userNameCreds[0].Key.Parent(), nil
}

func (s *AccountStore) GetAllAccounts(c context.Context, offset, limit int) ([]*Account, error) {
	var accounts []*Account
	doneChan := make(chan bool)
	errChan := make(chan error)
	dataChan := make(chan Account)

	keys, err := datastore.NewQuery(s.TableName).
		Limit(limit).
		Offset(offset).
		KeysOnly().
		GetAll(c, nil)
	if err != nil {
		return nil, err
	}
	for _, k := range keys {
		go func(key *datastore.Key) {
			var a Account
			if err := s.Get(c, key, &a); err != nil {
				if err != datastore.ErrNoSuchEntity {
					errChan <- err
				}
				return
			}
			dataChan <- a
		}(k)
	}
	async.New().Run(doneChan, errChan)

LOOP:
	for {
		select {
		case a := <-dataChan:
			accounts = append(accounts, &a)
		case <-doneChan:
			break LOOP
		case err = <-errChan:
			break LOOP
		case <-time.Tick(10 * time.Second):
			err = errors.New("GetAllAccount timeout")
			break LOOP
		}
	}

	return accounts, err
}
