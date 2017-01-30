package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/chrisolsen/ae/handler"
	"github.com/chrisolsen/aetemplate/core"
	"golang.org/x/net/context"
)

// SignupHandler .
type SignupHandler struct {
	handler.Base
}

func (h SignupHandler) ServeHTTP(c context.Context, w http.ResponseWriter, r *http.Request) {
	h.Bind(c, w, r)

	switch r.Method {
	case http.MethodPost:
		h.createAccount()
	case http.MethodOptions:
		h.ValidateOrigin([]string{"http://your_domain.com"})
	default:
		h.Abort(http.StatusNotFound, nil)
	}
}

// POST /v1/signup => [201, 400, 500]
//  {
//  	account: {
//  		firstName: "jim",
//  		...
//  	},
//  	credentials: {
//  		providerId: "234324523",
//  		providerName: "facebook",
//  		providerToken: "9q8763w4iwqr",
//
//			username: "bob@example.com",
// 			password: "foobario"
//  	}
//  }
func (h *SignupHandler) createAccount() {
	type data struct {
		Account     core.Account     `json:"account"`
		Credentials core.Credentials `json:"credentials"`
	}

	var input data
	err := json.NewDecoder(h.Req.Body).Decode(&input)
	if err != nil {
		h.Abort(http.StatusBadRequest, fmt.Errorf("decoding req body: %v", err))
		return
	}

	accountKey, err := AccountStore.Create(h.Ctx, &input.Credentials, &input.Account)
	if err != nil {
		h.Abort(http.StatusInternalServerError, fmt.Errorf("creating account: %v", err))
		return
	}

	token, err := TokenStore.Create(h.Ctx, accountKey)
	if err != nil {
		h.Abort(http.StatusInternalServerError, fmt.Errorf("creating token: %v", err))
		return
	}

	h.ToJSONWithStatus(token, http.StatusCreated)
}
