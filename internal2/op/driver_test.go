package op_test

import (
	"testing"

	_ "github.com/alist-org/alist/v3/drivers"
	"github.com/alist-org/alist/v3/internal2/op"
)

func TestDriverItemsMap(t *testing.T) {
	itemsMap := op.GetDriverInfoMap()
	if len(itemsMap) != 0 {
		t.Logf("driverInfoMap: %v", itemsMap)
	} else {
		t.Errorf("expected driverInfoMap not empty, but got empty")
	}
}
