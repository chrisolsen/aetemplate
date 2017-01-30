package images

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/chrisolsen/ae/handler"
	"github.com/chrisolsen/ae/image"
	"github.com/chrisolsen/ae/que"
	"golang.org/x/net/context"
)

func init() {
	q := que.New()

	http.Handle("/images.v1/images", q.Handle(ImageHandler{}))
	http.HandleFunc("/_ah/start", q.Then(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		// need an empty handler to allow the service to start up
	}))
}

type ImageHandler struct {
	handler.Base
}

func (h ImageHandler) ServeHTTP(c context.Context, w http.ResponseWriter, r *http.Request) {
	h.Bind(c, w, r)
	switch r.Method {
	case http.MethodGet:
		h.fetch()
	default:
		h.Abort(http.StatusNotFound, nil)
	}
}

// GET /images?name={foobar}&w={100}&h={100}
func (h *ImageHandler) fetch() {
	// query params
	name, ok := h.QueryParam("name")
	if !ok {
		h.Abort(http.StatusBadRequest, errors.New("name query param required"))
		return
	}

	var width, height int = 0, 0
	w, wok := h.QueryParam("w")
	if wok {
		width, _ = strconv.Atoi(w)
	}
	ht, hok := h.QueryParam("h")
	if hok {
		height, _ = strconv.Atoi(ht)
	}
	if !wok && !hok {
		h.Abort(http.StatusBadRequest, errors.New("width or height is required"))
		return
	}

	// fetch url for required size
	scheme := "https"
	if h.Req.TLS == nil {
		scheme = "http"
	}
	url, err := image.SizedURL(h.Ctx, scheme, name, width, height)
	if err != nil {
		h.Abort(http.StatusInternalServerError, fmt.Errorf("failed to get sized image: %v", err))
		return
	}
	http.Redirect(h.Res, h.Req, url, http.StatusMovedPermanently)
}
