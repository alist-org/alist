package data

import "github.com/alist-org/alist/v3/cmd/args"

func InitData() {
	initUser()
	initSettings()
	if args.Dev {
		initDevData()
	}
}
