package app

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/coachchris/core"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

const (
	tokensTable = "tokens"
)

// Errors
var (
	errMissingAuthToken   = errors.New("Auth token does not exist")
	errMissingAuthHeader  = errors.New("No authorization header supplied")
	errMultipleAuthTokens = errors.New("Duplicate auth token exist")
)

// Token keys
const (
	newTokenHeader       string = "new-auth-token"
	newTokenExpiryHeader string = "new-auth-token-expiry"
)

// TokenDetails is the data type that is stored in memcache using the token as a key.
type tokenDetails struct {
	Expiry     time.Time
	AccountKey string
	Token      string
}

func (t *tokenDetails) isExpired() bool {
	return t.Expiry.Before(time.Now())
}

func (t *tokenDetails) willExpireIn(duration time.Duration) bool {
	future := time.Now().Add(duration)
	return t.Expiry.Before(future)
}

type AuthMiddleware struct {
	tokenSvc core.TokenServicer
}

func (a *AuthMiddleware) ApiAuth(c context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	// let option requests through
	if r.Method == http.MethodOptions {
		return c
	}

	var err error
	c, cancel := context.WithCancel(c)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	tokenDetails, err := a.getTokenDetails(c, r.Header.Get("Authorization"))
	if err != nil {
		log.Errorf(c, "failed to get token details: %v", err)
		w.WriteHeader(http.StatusPreconditionFailed)
		return c
	}

	accountKey, err := datastore.DecodeKey(tokenDetails.AccountKey)
	if err != nil {
		log.Errorf(c, "failed to decode account key: %v", err)
		w.WriteHeader(http.StatusPreconditionFailed)
		return c
	}

	// add accountKey to context
	c = session.SetAccountKey(c, accountKey)

	// if token has expired return 401
	if tokenDetails.isExpired() {
		log.Errorf(c, "expired Token")
		w.WriteHeader(http.StatusUnauthorized)
		return c
	}

	// if the token's expiry less than a week away, get new token
	if tokenDetails.willExpireIn(time.Hour * 24 * 7) {
		newToken, err := a.getNewToken(c, accountKey)
		if err != nil {
			log.Errorf(c, "failed to create new token: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return c
		}

		// send back the new token values
		w.Header().Add(newTokenHeader, newToken.Value())
		w.Header().Add(newTokenExpiryHeader, newToken.Expiry.Format(time.RFC3339))
	}

	return c
}

// Gets the token for the rawToken value
func (a *AuthMiddleware) getTokenDetails(c context.Context, authHeader string) (*tokenDetails, error) {
	var err error

	if len(authHeader) <= len("token=") {
		return nil, errMissingAuthHeader
	}

	// prevent token caching with blank string value
	rawToken := authHeader[len("token="):]
	if len(rawToken) == 0 {
		return nil, errMissingAuthToken
	}

	tokenDetails, err := a.getCacheToken(c, rawToken)
	if err != nil && err != memcache.ErrCacheMiss {
		return nil, err
	}

	if err == memcache.ErrCacheMiss {
		tokenKey, err := datastore.DecodeKey(rawToken)
		if err != nil {
			return nil, fmt.Errorf("decoding token key: %v", err)
		}

		token, err := a.tokenSvc.Get(c, tokenKey)
		if err != nil {
			return nil, err
		}

		// no token found in the db either, user needs to sign in
		if token == nil {
			return nil, errMissingAuthToken
		}

		// add the token to memcache
		tokenDetails, err = a.setCacheToken(c, token.Key.Parent(), token)
		if err != nil {
			return nil, err
		}
	}

	return tokenDetails, nil
}

// getCacheToken attemps to fetch the token details for the raw token string passed in
func (a *AuthMiddleware) getCacheToken(c context.Context, rawToken string) (*tokenDetails, error) {
	var tokenDetails tokenDetails
	_, err := memcache.JSON.Get(c, rawToken, &tokenDetails)

	return &tokenDetails, err
}

// setCacheToken memcaches the passed in raw token value
func (a *AuthMiddleware) setCacheToken(c context.Context, accountKey *datastore.Key, token *core.Token) (*tokenDetails, error) {
	tokenDetails := tokenDetails{
		AccountKey: accountKey.Encode(),
		Expiry:     token.Expiry,
		Token:      token.Value(),
	}

	// save to memcache
	err := memcache.JSON.Set(c, &memcache.Item{
		Key:        token.Value(),
		Object:     tokenDetails,
		Expiration: -1 * time.Since(token.Expiry),
	})
	if err != nil {
		return nil, err
	}

	return &tokenDetails, nil
}

// Creates a new token and links it to the account for the old token
func (a *AuthMiddleware) getNewToken(c context.Context, accountKey *datastore.Key) (*core.Token, error) {
	if accountKey == nil {
		return nil, errors.New("account key is required to create a token")
	}

	token := a.tokenSvc.NewToken()
	_, err := a.tokenSvc.AddToken(c, accountKey, token)

	return token, err
}
