package drivers

import (
	_ "github.com/alist-org/alist/v3/drivers/local"
	_ "github.com/alist-org/alist/v3/drivers/onedrive"
	_ "github.com/alist-org/alist/v3/drivers/virtual"
)

// All do nothing,just for import
// same as _ import
func All() {

}
