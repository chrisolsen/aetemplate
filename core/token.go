package core

import (
	"time"

	"github.com/chrisolsen/aestore"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type Token struct {
	Model
	Expiry time.Time `json:"expiry" datastore:",noindex"`
}

// Value returns the datastore key's string value
func (t *Token) Value() string {
	return t.Key.Encode()
}

type tokenStore struct {
	aestore.Base
}

func newTokenStore() tokenStore {
	s := tokenStore{}
	s.TableName = "tokens"
	return s
}

type TokenServicer interface {
	Get(c context.Context, key *datastore.Key) (*Token, error)
	AddToken(c context.Context, accountKey *datastore.Key, t *Token) (*datastore.Key, error)
	NewToken() *Token
}

type TokenService struct{}

func NewTokenService() TokenService {
	return TokenService{}
}

func (s TokenService) Get(c context.Context, key *datastore.Key) (*Token, error) {
	var t Token
	store := newTokenStore()
	err := store.Get(c, key, &t)
	return &t, err
}

func (s TokenService) AddToken(c context.Context, accountKey *datastore.Key, t *Token) (*datastore.Key, error) {
	store := newTokenStore()
	return store.Create(c, t, accountKey)
}

func (s TokenService) NewToken() *Token {
	return &Token{Expiry: time.Now().AddDate(0, 2, 0)}
}
