package app

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/chrisolsen/ae/handler"
	"github.com/chrisolsen/aetemplate/core"
	"golang.org/x/net/context"
)

type AuthHandler struct {
	handler.Base
}

func (h AuthHandler) ServeHTTP(c context.Context, w http.ResponseWriter, r *http.Request) {
	h.Bind(c, w, r)
	svc := core.AuthService{
		URLGetter: core.AppEngineURLGetter{Ctx: c},
	}
	switch r.Method {
	case http.MethodPost:
		h.authenticateUser(svc.Authenticate)
	case http.MethodOptions:
		h.ValidateOrigin(nil)
	default:
		h.Abort(http.StatusNotFound, nil)
	}
}

// Authenticate authenticates the submitted credentials and returns an auth token
// for the found acccount in the response. The credentials can either include
// authh provider details or email/password
//
// 	200 - authenticated
// 	401 - not authenticated
//  400 - bad request
//
// 	POST /v1/auth
//	{
// 		"accountKey": "12352345",  			// passed in on initial signup only
//  	"providerId": 21234234,
//  	"providerName": "facebook",
//  	"providerToken": "8a7wi2jrhfas...",
//
// 		"email": "john@example.com",
// 		"password": "foobar"
//  }
func (h *AuthHandler) authenticateUser(authenticate core.AuthFunc) {
	var creds core.Credentials
	err := json.NewDecoder(h.Req.Body).Decode(&creds)
	if err != nil {
		h.Abort(http.StatusBadRequest, err)
		return
	}

	if !creds.Valid() {
		h.Abort(http.StatusBadRequest, errors.New("missing required credentials"))
		return
	}

	token, err := authenticate(h.Ctx, &creds)
	if err != nil {
		h.Abort(http.StatusUnauthorized, err)
	}

	h.ToJSON(token)
}
