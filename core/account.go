package core

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chrisolsen/aestore"
	"github.com/chrisolsen/async"
	"golang.org/x/crypto/bcrypt"
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
	Model

	FirstName string `json:"firstName" datastore:",noindex"`
	LastName  string `json:"lastName" datastore:",noindex"`
	Gender    string `json:"gender" datastore:",noindex"`
	Locale    string `json:"locale" datastore:",noindex"`
	Location  string `json:"location" datastore:",noindex"`
	Name      string `json:"name" datastore:",noindex"`
	Timezone  int    `json:"timezone" datastore:",noindex"`
	Email     string `json:"email"`

	Photo Attachment `json:"photo"`

	// lowercased attributes for searches
	FirstNameFilter string `json:"-"`
	LastNameFilter  string `json:"-"`
	NameFilter      string `json:"-"`
}

// Load - PropertyLoadSaver interface
func (a *Account) Load(ps []datastore.Property) error {
	if err := datastore.LoadStruct(a, ps); err != nil {
		switch err.(type) {
		case *datastore.ErrFieldMismatch:
			return nil
		default:
			return err
		}
	}
	return nil
}

// Save - PropertyLoadSaver interface
func (a *Account) Save() ([]datastore.Property, error) {
	a.FirstNameFilter = strings.ToLower(a.FirstName)
	a.LastNameFilter = strings.ToLower(a.LastName)
	a.NameFilter = strings.ToLower(a.Name)
	return datastore.SaveStruct(a)
}

// AccountService contains logic methods called on from the api handlers
type AccountService struct{}

// EncryptPassword converts the raw password to a brcypt hash
func (s *AccountService) EncryptPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(b), err
}

// ValidatePassword checks that the saved hash and raw password hash match
func (s *AccountService) ValidatePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// GetAllAccounts fetches a paginated list of all the accounts
func (s *AccountService) GetAllAccounts(c context.Context, offset, limit int) ([]*Account, error) {
	store := accountStore{}
	return store.GetAllAccounts(c, offset, limit)
}

// GetAccountKeyByCredentials fetches the account matching the auth provider credentials
func (s *AccountService) GetAccountKeyByCredentials(c context.Context, creds *Credentials) (*datastore.Key, error) {
	var err error
	cstore := credentialStore{}
	// on initial signup the account key will exist within the credentials
	if creds.AccountKey != nil {
		var accountCreds []*Credentials
		_, err = cstore.GetByParent(c, creds.AccountKey, &accountCreds, nil)
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

	if len(creds.ProviderID) > 0 {
		return cstore.GetAccountKeyByProvider(c, creds)
	}
	}
	return store.GetAccountKeyByEmailAndPassword(c, creds.Username, creds.Password)
}

type accountStore struct {
	aestore.Base
}

func newAccountStore() accountStore {
	s := accountStore{}
	s.TableName = "accounts"
	return s
}

func (s *accountStore) GetAllAccounts(c context.Context, offset, limit int) ([]*Account, error) {
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
				errChan <- err
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
