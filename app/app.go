package app

import (
	"net/http"

	"github.com/chrisolsen/ae/handler"
	"github.com/chrisolsen/ae/que"
	Session "github.com/chrisolsen/ae/session"
	"github.com/chrisolsen/aetemplate/core"
	"golang.org/x/net/context"
)

var (
	session = Session.New(Session.CacheTypeGob, 0)
)

var (
	AccountStore    = core.NewAccountStore()
	AttachmentStore = core.AttachmentStore{}
	CredentialStore = core.NewCredentialStore()
	TokenStore      = core.NewTokenStore()
)

var (
	authMiddleware = AuthMiddleware{}
	// set text/json response type
	jsonMiddleware = func(c context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		w.Header().Add("Content-Type", "text/json; charset=utf-8")
		return c
	}
)

func init() {
	// no auth
	noAuth := que.New(handler.OriginMiddleware(nil))
	http.Handle("/v1/auth", noAuth.Handle(AuthHandler{}))
	http.Handle("/v1/signup", noAuth.Handle(SignupHandler{}))

	// auth
	auth := que.New(handler.OriginMiddleware(nil), authMiddleware.APIAuth)
	http.Handle("/v1/me", auth.Handle(AccountsHandler{}))

	// static files
	http.Handle("/static/", http.FileServer(http.Dir("static")))

	// ex. mail auth handler
	http.HandleFunc("/.well-known/acme-challenge/qdKxHj8XnWKq91pOctds5wzZUrW7TWTH4NbmwH5oa_g", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("qdKxHj8XnWKq91pOctds5wzZUrW7TWTH4NbmwH5oa_g.7RL12VvymIwJ6NoXGAimP_CgJQ7JrNQEjfufjLxKRiQ"))
	})
}
