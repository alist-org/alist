package operations_test

import (
	"testing"

	_ "github.com/alist-org/alist/v3/drivers"
	"github.com/alist-org/alist/v3/internal/operations"
)

func TestDriverItemsMap(t *testing.T) {
	itemsMap := operations.GetDriverItemsMap()
	if len(itemsMap) != 0 {
		t.Logf("driverItemsMap: %v", itemsMap)
	} else {
		t.Errorf("expected driverItemsMap not empty, but got empty")
	}
}
