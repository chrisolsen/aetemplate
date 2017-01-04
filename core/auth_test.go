package core

import (
	"testing"

	"google.golang.org/appengine/datastore"
)

func TestEndpoints_Auth(t *testing.T) {
	c := getContext()

	type signupTest struct {
		name        string
		creds       *Credentials
		expectedErr error
	}

	var tests = []signupTest{
		{
			name:        "Invalid account",
			creds:       &Credentials{ProviderName: "facebook", ProviderID: "1234", ProviderToken: "asoiudykaejhes"},
			expectedErr: nil,
		},
	}

	// account to auth with
	a := Account{}
	pkey, _ := datastore.Put(c, datastore.NewIncompleteKey(c, "accounts", nil), &a)
	creds := Credentials{ProviderID: "1234", ProviderName: "facebook", ProviderToken: "asoiudykaejhes"}
	ckey, _ := datastore.Put(c, datastore.NewIncompleteKey(c, "credentials", pkey), &creds)

	for _, ts := range tests {
		ts.creds.Key = ckey
		func(test signupTest) {
			authService := AuthService{
				URLGetter: mockURLGetter{err: test.expectedErr, body: `{"id": "1234"}`},
			}

			token, err := authService.Authenticate(c, test.creds)
			if err != nil && test.expectedErr == nil {
				t.Error("Unexpected error", err.Error())
				return
			}

			if err != nil && test.expectedErr != nil {
				return
			}

			if len(token.Value()) == 0 {
				t.Error("No token returned")
				return
			}
		}(ts)
	}
}
