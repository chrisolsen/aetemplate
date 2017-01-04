package core

import (
	"errors"
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
	// when creds are first created `key` will exist. Accessing them by key ensures we
	// can fetch the credentials if they are not yet available on all dataservers
	if creds.Key != nil {
		return creds.Key.Parent(), nil
	}

	store := newCredentialStore()
	if len(creds.ProviderToken) > 0 {
		return store.GetAccountKeyByProviderId(c, creds.ProviderID)
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
