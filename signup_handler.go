package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/chrisolsen/aehandler"
	"gitlab.com/coachchris/core"
	"golang.org/x/net/context"
)

type SignupHandler struct {
	aehandler.Base
}

func (h SignupHandler) ServeHTTP(c context.Context, w http.ResponseWriter, r *http.Request) {
	h.Bind(c, w, r)
	svc := core.AccountService{}

	switch r.Method {
	case http.MethodPost:
		h.createAccount(svc.Create)
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
func (h *SignupHandler) createAccount(createAccount core.AccountCreateFunc) {
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

	input.Account.Key, err = createAccount(h.Ctx, &input.Credentials, &input.Account)
	if err != nil {
		h.Abort(http.StatusInternalServerError, fmt.Errorf("creating account: %v", err))
		return
	}

	h.ToJSONWithStatus(input.Account, http.StatusCreated)
}
