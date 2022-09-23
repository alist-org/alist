package bootstrap

import (
	"github.com/alist-org/alist/v3/internal/aria2"
	"github.com/alist-org/alist/v3/pkg/utils"
)

func InitAria2() {
	go func() {
		_, err := aria2.InitClient(2)
		if err != nil {
			//utils.Log.Errorf("failed to init aria2 client: %+v", err)
			utils.Log.Infof("Aria2 not ready.")
		}
	}()
}
