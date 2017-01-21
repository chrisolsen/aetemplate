package core

import (
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
)

type URLGetter interface {
	Get(url string) (*http.Response, error)
}

type AppEngineURLGetter struct {
	Ctx context.Context
}

func (ug AppEngineURLGetter) Get(url string) (*http.Response, error) {
	client := urlfetch.Client(ug.Ctx)
	return client.Get(url)
}
