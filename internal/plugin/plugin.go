package plugin

import (
	"strings"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/generic_sync"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"

	plugin_manage "github.com/alist-org/alist/v3/internal/plugin/manage"
	_ "github.com/alist-org/alist/v3/internal/plugin/yaegi"
)

var pluginControlMap = generic_sync.MapOf[string, model.PluginControl]{}

func InitPlugin() error {
	plugins, err := db.GetEnabledPlugin()
	if err != nil {
		return err
	}
	var errs []error
	for _, plugin := range plugins {
		if _, err := RegisterPlugin(plugin); err != nil {
			errs = append(errs, err)
		}
	}
	return utils.MergeErrors(errs...)
}

func AddPlugin(plugin model.Plugin) error {
	adp, err := RegisterPlugin(plugin)
	if err != nil {
		return err
	}

	pluginInfo, err := adp.Config()
	if err != nil {
		return utils.MergeErrors(err, UnRegisterPlugin(plugin))
	}

	plugin.UUID = pluginInfo.GetUUID()
	plugin.Type = strings.Join(pluginInfo.GetType(), ",")
	plugin.Version = pluginInfo.GetVersion()
	plugin.ApiVersion = strings.Join(pluginInfo.GetApiVersion(), ",")
	if err = db.CreatePlugin(&plugin); err != nil {
		return utils.MergeErrors(err, UnRegisterPlugin(plugin))
	}
	return nil
}

func DeletePluginByID(id uint) error {
	dbPlugin, err := db.GetPluginById(id)
	if err != nil {
		return errors.WithMessage(err, "failed get plugin")
	}

	err = db.DeleteStorageById(dbPlugin.ID)
	if err != nil {
		return errors.WithMessage(err, "failed delete storage in database")
	}
	return utils.MergeErrors(UnRegisterPlugin(*dbPlugin), err)
}

func DisablePluginByID(id uint) error {
	dbPlugin, err := db.GetPluginById(id)
	if err != nil {
		return errors.WithMessage(err, "failed get plugin")
	}

	dbPlugin.Disabled = true
	return utils.MergeErrors(UnRegisterPlugin(*dbPlugin), db.UpdatePlugin(dbPlugin))
}

func EnablePluginByID(id uint) error {
	dbPlugin, err := db.GetPluginById(id)
	if err != nil {
		return errors.WithMessage(err, "failed get plugin")
	}
	if dbPlugin.Disabled {
		if _, err := RegisterPlugin(*dbPlugin); err != nil {
			return err
		}
	}

	dbPlugin.Disabled = false
	return db.UpdatePlugin(dbPlugin)
}

func RegisterPlugin(plugin model.Plugin) (model.PluginControl, error) {
	if !utils.SliceContains(strings.Split(plugin.ApiVersion, ","), "v1") {
		return nil, errors.Errorf("only support plugin api version is v1")
	}

	if !pluginControlMap.Has(plugin.UUID) {
		if new, ok := plugin_manage.GetPluginControlManage(plugin.Mode); ok {
			p, err := new(plugin)
			if err != nil {
				return nil, err
			}

			_, err = p.Config()
			if err != nil {
				return nil, errors.Errorf("unable to obtain plugin info, err: %w", err)
			}

			if err = p.Load(); err != nil {
				return nil, err
			}
			pluginControlMap.Store(plugin.UUID, p)
			return p, nil
		}
		return nil, errors.Errorf("not support plugin mode: %s", plugin.Mode)
	}
	return nil, errors.New("the plugin has been loaded")
}

func UnRegisterPlugin(plugin model.Plugin) error {
	if adp, ok := pluginControlMap.Load(plugin.UUID); ok {
		pluginControlMap.Delete(plugin.UUID)
		return utils.MergeErrors(adp.Unload(), adp.Release())
	}
	return nil
}
