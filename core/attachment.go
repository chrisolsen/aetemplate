package core

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/chrisolsen/ae/image"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
)

// Attachment links data saved in an external storage
type Attachment struct {
	Name string `json:"name"`
	Type string `json:"type"`

	// base64 encoded data passed up from client
	Data string `json:"data,omitempty" datastore:"-"`
}

// Bytes trims the meta data from the encoded string and converts the data to []byte
func (ra *Attachment) Bytes() ([]byte, error) {
	index := strings.Index(ra.Data, ",") + 1
	data, err := base64.StdEncoding.DecodeString(ra.Data[index:])
	return []byte(data), err
}

// AttachmentStore provides the methods to save to the external storage service
type AttachmentStore struct{}

// AttachmentStorer makes testing easier
type AttachmentStorer interface {
	CreateWithData(c context.Context, data []byte, contentType string) (*Attachment, error)
	CreateWithURL(c context.Context, url string) (*Attachment, error)
}

// CreateWithData saves the passed in data as an attachment
func (as AttachmentStore) CreateWithData(c context.Context, data []byte, contentType string) (*Attachment, error) {
	name := uuid.NewV4().String()

	// save image
	writer, err := image.NewWriter(c, name, contentType)
	if err != nil {
		return nil, fmt.Errorf("creating image writer: %v", err)
	}
	count, err := writer.Write(data)
	if count <= 0 {
		return nil, errors.New("zero bytes written for image")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to save image to storage: %v", err)
	}
	defer writer.Close()

	return &Attachment{Name: name, Type: contentType}, nil
}

// CreateWithURL performs an external fetch of the data with the URL and saves
// the returned data as an attachment
func (as AttachmentStore) CreateWithURL(c context.Context, url string) (*Attachment, error) {
	// get image
	client := urlfetch.Client(c)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get image with URL: %v", err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	return as.CreateWithData(c, data, resp.Header.Get("Content-Type"))
}
