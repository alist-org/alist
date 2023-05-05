package plugin_manage

import "github.com/alist-org/alist/v3/internal/model"

type PluginControlManage func(string) (model.PluginControl, error)

var pluginControlManageMap = map[string]PluginControlManage{}

// 注册插件适配器
func RegisterPluginControlManage(name string, new PluginControlManage) {
	pluginControlManageMap[name] = new
}

// 获取插件适配器
func GetPluginControlManage(name string) (PluginControlManage, bool) {
	new, ok := pluginControlManageMap[name]
	return new, ok
}
