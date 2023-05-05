package plugin_manage

import (
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/generic_sync"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
)

var pluginControlMap = generic_sync.MapOf[string, model.PluginControl]{}

// 注册插件
func RegisterPlugin(plugin model.Plugin) (model.PluginControl, error) {
	if !pluginControlMap.Has(plugin.UUID) {
		if new, ok := GetPluginControlManage(plugin.Mode); ok {
			// 初始化插件控件
			adp, err := new(plugin.Path)
			if err != nil {
				return nil, err
			}

			// 从插件中读取信息保存到数据库
			pluginDB, err := FixPluginConfig(adp, plugin)
			if err != nil {
				if err2 := adp.Release(); err2 != nil {
					return nil, utils.MergeErrors(err, err2)
				}
				return nil, err
			}

			// 是否支持插件ApiVersion
			if !IsSupportPlugin(pluginDB.ApiVersion) {
				err := errors.Errorf("only support plugin api version is %s,but plugin api version is %s", PLUGIN_API_VERSION.String(), pluginDB.ApiVersion)
				if err2 := adp.Release(); err2 != nil {
					return nil, utils.MergeErrors(err, err2)
				}
				return nil, err
			}

			// 加载插件
			if err = adp.Load(); err != nil {
				if err2 := adp.Release(); err2 != nil {
					return nil, utils.MergeErrors(err, err2)
				}
				return nil, err
			}
			pluginControlMap.Store(pluginDB.UUID, adp)
			return adp, nil
		}
		return nil, errors.Errorf("not support plugin mode: %s", plugin.Mode)
	}
	return nil, errs.PluginHasBeenLoaded
}

// 卸载插件
func UnRegisterPlugin(plugin model.Plugin) error {
	if adp, ok := pluginControlMap.Load(plugin.UUID); ok {
		if err := adp.Unload(); err != nil {
			return err
		}
		if err := adp.Release(); err != nil {
			return err
		}
		pluginControlMap.Delete(plugin.UUID)
	}
	return nil
}
