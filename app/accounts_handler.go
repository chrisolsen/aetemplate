package app

import (
	"fmt"
	"net/http"

	"github.com/chrisolsen/ae/handler"
	"github.com/chrisolsen/aetemplate/core"
	"golang.org/x/net/context"
)

type AccountsHandler struct {
	handler.Base
}

func (h AccountsHandler) ServeHTTP(c context.Context, w http.ResponseWriter, r *http.Request) {
	h.Bind(c, w, r)
	switch r.Method {
	case http.MethodGet:
		h.getMe()
	case http.MethodOptions:
		h.ValidateOrigin(nil)
	default:
		h.Abort(http.StatusNotFound, nil)
	}
}

func (h *AccountsHandler) getMe() {
	var me core.Account
	err := session.Account(h.Ctx, &me)
	if err != nil {
		h.Abort(http.StatusBadRequest, fmt.Errorf("unable to get account from token: %v", err))
		return
	}

	h.ToJSON(&me)
}
