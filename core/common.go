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
	ctx context.Context
}

func (ug AppEngineURLGetter) Get(url string) (*http.Response, error) {
	client := urlfetch.Client(ug.ctx)
	return client.Get(url)
}
