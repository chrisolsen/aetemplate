package app

import (
	"net/http"

	"github.com/chrisolsen/aehandler"
	"golang.org/x/net/context"
)

type helloHandler struct {
	aehandler.Base
}

func (h helloHandler) ServeHTTP(c context.Context, w http.ResponseWriter, r *http.Request) {
	h.Bind(c, w, r)
	switch r.Method {
	case http.MethodGet:
		h.sayHello()
	case http.MethodOptions:
		h.ValidateOrigin([]string{"https://some.origin.com"})
	default:
		h.Abort(http.StatusNotFound, nil)
	}
}

// GET /hello
//  {
//      message: "hello"
//  }
func (h *helloHandler) sayHello() {
	type hello struct {
		Message string `json:"message"`
	}

	h.ToJSON(hello{Message: "Hello, World"})
}
