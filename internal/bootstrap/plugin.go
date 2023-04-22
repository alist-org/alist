package bootstrap

import (
	"github.com/alist-org/alist/v3/internal/plugin"
	"github.com/alist-org/alist/v3/pkg/utils"
)

func InitPlugin() error {
	utils.Log.Info("init Plugin")
	return plugin.InitPlugin()
}
