package drivers

import (
	_ "github.com/alist-org/alist/v3/drivers/123"
	_ "github.com/alist-org/alist/v3/drivers/aliyundrive"
	_ "github.com/alist-org/alist/v3/drivers/baidu_netdisk"
	_ "github.com/alist-org/alist/v3/drivers/google_drive"
	_ "github.com/alist-org/alist/v3/drivers/local"
	_ "github.com/alist-org/alist/v3/drivers/onedrive"
	_ "github.com/alist-org/alist/v3/drivers/pikpak"
	_ "github.com/alist-org/alist/v3/drivers/quark"
	_ "github.com/alist-org/alist/v3/drivers/teambition"
	_ "github.com/alist-org/alist/v3/drivers/virtual"
)

// All do nothing,just for import
// same as _ import
func All() {

}
