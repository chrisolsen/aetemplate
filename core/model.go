package core

import "google.golang.org/appengine/datastore"

// Model has the common key property
type Model struct {
	Key *datastore.Key `json:"key" datastore:"-"`
}

// SetKey allows the key of a model to be auto set
func (b *Model) SetKey(k *datastore.Key) {
	b.Key = k
}
