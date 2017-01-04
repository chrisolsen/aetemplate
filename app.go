package app

import (
	"net/http"
	"os"
	"strings"

	"gitlab.com/coachchris/core"

	"github.com/chrisolsen/aehandler"
	"github.com/chrisolsen/aesession"
	"github.com/chrisolsen/quincy"
	"golang.org/x/net/context"
)

var (
	session = aesession.New(aesession.CacheTypeGob, 0)
)

var (
	// authenticates user
	authMiddleware = AuthMiddleware{tokenSvc: core.NewTokenService()}

	// set text/json response type
	jsonMiddleware = func(c context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		w.Header().Add("Content-Type", "text/json; charset=utf-8")
		return c
	}
)

var allowedOrigins []string

func init() {
	// pull values from app.yaml file
	if len(os.Getenv("ALLOWED_ORIGINS")) > 0 {
		allowedOrigins = strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	}

	// middleware
	auth := quincy.New(authMiddleware.ApiAuth, aehandler.OriginMiddleware(allowedOrigins), jsonMiddleware)
	noAuth := quincy.New(aehandler.OriginMiddleware(allowedOrigins), jsonMiddleware)

	// no auth
	http.Handle("/v1/hello", noAuth.Handle(helloHandler{}))
	http.Handle("/v1/auth", noAuth.Handle(AuthHandler{}))
	http.Handle("/v1/signup", noAuth.Handle(SignupHandler{}))

	// auth
	http.Handle("/v1/attachments", auth.Handle(AttachmentHandler{}))
}
