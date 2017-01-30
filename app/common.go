package app

import (
	"github.com/chrisolsen/ae/store"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type modelGetter interface {
	Get(c context.Context, key *datastore.Key, dst store.Model) error
}

type modelUpdater interface {
	Update(c context.Context, key *datastore.Key, model interface{}) error
}

type modelCreater interface {
	Create(c context.Context, model store.Model, parent *datastore.Key) (*datastore.Key, error)
}

type modelCopier interface {
	Copy(c context.Context, srcKey *datastore.Key, dst store.Model) error
}

type modelDeleter interface {
	Delete(c context.Context, key *datastore.Key) error
}
