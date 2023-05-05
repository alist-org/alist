package plugin_manage

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"gorm.io/gorm"
)

// 将不在数据库中的插件加入数据库
func AddLocalPluginToDB() error {
	plugins, err := scanLocalPlugin()
	if err != nil {
		return err
	}

	var errs []error
	for _, plugin := range plugins {
		if _, err := db.GetPluginByModeAndPath(plugin.Mode, plugin.Path); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				plugin.Disabled = true
				if _, err := FixPluginConfigByModel(plugin); err != nil {
					errs = append(errs, err)
				}
				continue
			}
			errs = append(errs, err)
		}
	}
	return utils.MergeErrors(errs...)
}

// 扫描插件目录，获取所有可识别插件
func scanLocalPlugin() ([]model.Plugin, error) {
	pluginDir := conf.Conf.PluginDir
	var plugins []model.Plugin
	err := filepath.WalkDir(pluginDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		targpath, _ := filepath.Rel(pluginDir, path)
		names := strings.Split(targpath, string(filepath.Separator))
		if names[0] == "src" && names[len(names)-1] == "plugin.go" {
			plugins = append(plugins, model.Plugin{
				Mode: model.PLUGIN_MODE_YAEGI,
				Path: strings.Join(names[1:len(names)-1], "/"),
			})
			return filepath.SkipDir
		}
		// plugin.go 应该存在于第一层目录
		if len(names) > 3 {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return plugins, nil
}

func FixPluginConfigByPath(path string) (*model.Plugin, error) {
	mode := CheckPluginMode(path)
	if new, ok := GetPluginControlManage(mode); ok {
		adp, err := new(path)
		if err != nil {
			return nil, err
		}
		defer adp.Release()
		return FixPluginConfig(adp, model.Plugin{Mode: mode, Path: path})
	}
	return nil, errs.NotSupportPluginMode
}

func FixPluginConfigByModel(plugin model.Plugin) (*model.Plugin, error) {
	if new, ok := GetPluginControlManage(plugin.Mode); ok {
		adp, err := new(plugin.Path)
		if err != nil {
			return nil, err
		}
		defer adp.Release()
		return FixPluginConfig(adp, plugin)
	}
	return nil, errs.NotSupportPluginMode
}

// 修复插件在数据库中的信息,如果不存在则创建
// model.Plugin 继承原本的 {ID,Mode,Path,Disabled},其余部分从adp中获取
func FixPluginConfig(adp model.PluginControl, plugin model.Plugin) (*model.Plugin, error) {
	config, err := adp.Config()
	if err != nil {
		return nil, err
	}

	newPlugin := model.Plugin{
		ID:       plugin.ID,
		Mode:     plugin.Mode,
		Path:     plugin.Path,
		Disabled: plugin.Disabled,

		Name:       config.GetName(),
		UUID:       config.GetUUID(),
		Type:       strings.Join(config.GetType(), ","),
		Version:    config.GetVersion(),
		ApiVersion: strings.Join(config.GetApiVersion(), ","),
	}
	if newPlugin != plugin {
		if err := db.UpdatePlugin(&newPlugin); err != nil {
			return nil, err
		}
	}
	return &newPlugin, nil
}
