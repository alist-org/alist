package bootstrap

import "github.com/alist-org/alist/v3/internal/aria2"

func InitAria2() {
	go func() {
		_ = aria2.InitClient(2)
	}()
}
