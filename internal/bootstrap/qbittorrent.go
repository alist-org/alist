package bootstrap

import (
	"github.com/alist-org/alist/v3/internal/qbittorrent"
	"github.com/alist-org/alist/v3/pkg/utils"
)

func InitQbittorrent() {
	go func() {
		err := qbittorrent.InitClient()
		if err != nil {
			utils.Log.Infof("qbittorrent not ready.")
		}
	}()
}
