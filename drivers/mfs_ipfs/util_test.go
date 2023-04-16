package mfs_ipfs_test

import (
	"os"
	"testing"

	"github.com/alist-org/alist/v3/drivers/mfs_ipfs/util"
)

type TestRefresh struct{ id string }

func (r *TestRefresh) Get() (id string) { return r.id }
func (r *TestRefresh) Set(id string)    { r.id = id }
func TestUtil(t *testing.T) {
	util.DefaultPath = os.TempDir()
	if mapi, err := util.NewMfs("", "", &TestRefresh{id: "bafybeiczsscdsbs7ffqz55asqdf3smv6klcw3gofszvwlyarci47bgf354"}, nil, nil); err == nil {
		defer func() {
			if err = mapi.Close(); err != nil {
				t.Error(err)
			}
		}()
		if _, err := mapi.List(""); err != nil {
			t.Error(err)
		}
	} else {
		t.Error(err)
	}
}
