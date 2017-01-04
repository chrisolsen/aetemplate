package core

import (
	"os"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

var _inst aetest.Instance

func TestMain(m *testing.M) {
	_inst, _ = aetest.NewInstance(nil)
	os.Exit(func() int {
		id := m.Run()
		if _inst != nil {
			_inst.Close()
		}
		return id
	}())
}

func getContext() context.Context {
	inst := getInstance()
	r, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		inst.Close()
		return nil
	}
	return appengine.NewContext(r)
}

func getInstance() aetest.Instance {
	return _inst
}
