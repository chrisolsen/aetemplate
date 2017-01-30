package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/chrisolsen/ae/handler"
	"github.com/chrisolsen/aetemplate/core"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type AttachmentHandler struct {
	handler.Base
}

func (h AttachmentHandler) ServeHTTP(c context.Context, w http.ResponseWriter, r *http.Request) {
	h.Bind(c, w, r)

	switch r.Method {
	case http.MethodPost:
		h.create()
	case http.MethodOptions:
		h.ValidateOrigin([]string{"https://your_app.com"})
	default:
		h.Abort(http.StatusNotFound, nil)
	}
}

type AttachmentBody struct {
	URL         string `json:"url"`
	Data        []byte `json:"data"`
	ContentType string `json:"contentType"`
}

// POST /attachments?parent={key}
func (h *AttachmentHandler) create() {
	var body AttachmentBody
	var err error
	svc := core.AttachmentStore{}

	err = json.NewDecoder(h.Req.Body).Decode(&body)
	if err != nil {
		h.Abort(http.StatusBadRequest, fmt.Errorf("decode json: %v", err))
		return
	}

	parentKey, ok := h.QueryKey("parent")
	if !ok {
		h.Abort(http.StatusBadRequest, errors.New("invalid parent querystring key"))
		return
	}

	// create attachment
	var attachment *core.Attachment
	if len(body.URL) > 0 {
		attachment, err = svc.CreateWithURL(h.Ctx, body.URL)
	} else {
		attachment, err = svc.CreateWithData(h.Ctx, body.Data, body.ContentType)
	}
	if err != nil {
		h.Abort(http.StatusBadRequest, fmt.Errorf("failed finding parent: %v", err))
		return
	}

	// associate to parent
	switch parentKey.Kind() {
	case "accounts":
		err = h.associateToAccount(parentKey, attachment)
	default:
		h.Abort(http.StatusBadRequest, errors.New("unhandled attachment parent"))
	}
	if err != nil {
		h.Abort(http.StatusInternalServerError, fmt.Errorf("failed to create attachment: %v", err))
		return
	}

	h.ToJSONWithStatus(attachment, http.StatusCreated)
}

func (h *AttachmentHandler) associateToAccount(key *datastore.Key, photo *core.Attachment) error {
	var a core.Account
	err := AccountStore.Get(h.Ctx, key, &a)
	if err != nil {
		return fmt.Errorf("getting account: %v", err)
	}
	a.Photo = *photo
	err = AccountStore.Update(h.Ctx, a.Key, &a)
	if err != nil {
		return fmt.Errorf("updating account: %v", err)
	}

	return nil
}
