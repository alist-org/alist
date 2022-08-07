package data

import "github.com/alist-org/alist/v3/cmd/flags"

func InitData() {
	initUser()
	initSettings()
	if flags.Dev {
		initDevData()
		initDevDo()
	}
}
