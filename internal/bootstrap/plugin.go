package bootstrap

import (
	"github.com/alist-org/alist/v3/internal/plugin"
	"github.com/alist-org/alist/v3/pkg/utils"
)

func InitPlugin() error {
	utils.Log.Info("init Plugin")
	err := plugin.InitPlugin()
	if err != nil {
		utils.Log.Errorf("load plugin err: %s", err)
		return err
	}
	return nil
}
