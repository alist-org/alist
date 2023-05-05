package plugin

import (
	"context"
	"path/filepath"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/singleflight"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"

	plugin_manage "github.com/alist-org/alist/v3/internal/plugin/manage"
	_ "github.com/alist-org/alist/v3/internal/plugin/yaegi"
)

func InitPlugin() (err error) {
	// abs
	if !filepath.IsAbs(conf.Conf.PluginDir) {
		conf.Conf.PluginDir, err = filepath.Abs(conf.Conf.PluginDir)
		if err != nil {
			return err
		}
	}

	// scan add local plugin
	if err := plugin_manage.AddLocalPluginToDB(); err != nil {
		utils.Log.Warn(err)
	}

	// load
	plugins, err := db.GetEnabledPlugin()
	if err != nil {
		return err
	}

	var errs []error
	for _, plugin := range plugins {
		if _, err := plugin_manage.RegisterPlugin(plugin); err != nil {
			errs = append(errs, err)
		}
	}
	return utils.MergeErrors(errs...)
}

func DisablePluginByID(ctx context.Context, id uint) error {
	dbPlugin, err := db.GetPluginById(id)
	if err != nil {
		return errors.WithMessage(err, "failed get plugin")
	}
	if !dbPlugin.Disabled {
		if err := plugin_manage.UnRegisterPlugin(*dbPlugin); err != nil {
			return err
		}
	}

	dbPlugin.Disabled = true
	return db.UpdatePlugin(dbPlugin)
}

func EnablePluginByID(ctx context.Context, id uint) error {
	dbPlugin, err := db.GetPluginById(id)
	if err != nil {
		return errors.WithMessage(err, "failed get plugin")
	}
	if dbPlugin.Disabled {
		if _, err := plugin_manage.RegisterPlugin(*dbPlugin); err != nil {
			return err
		}
	}

	dbPlugin.Disabled = false
	return db.UpdatePlugin(dbPlugin)
}

func GetPluginRepository(ctx context.Context) []model.PluginInfo {
	return plugin_manage.GetAllPluginRepository()
}

func UpdatePluginRepository(ctx context.Context) error {
	return plugin_manage.UpdatePluginRepository(ctx)
}

var installG singleflight.Group[*model.Plugin]

func InstallPlugin(ctx context.Context, uuid string, version string) (*model.Plugin, error) {
	plugin, err, _ := installG.Do(uuid, func() (*model.Plugin, error) {
		plugin, err := plugin_manage.InstallPlugin(ctx, uuid, version)
		if err != nil {
			return nil, err
		}
		_, err = plugin_manage.RegisterPlugin(*plugin)
		return plugin, err
	})
	return plugin, err
}

func UninstallPlugin(ctx context.Context, id uint) error {
	dbPlugin, err := db.GetPluginById(id)
	if err != nil {
		return errors.WithMessage(err, "failed get plugin")
	}
	if err := plugin_manage.UnRegisterPlugin(*dbPlugin); err != nil {
		return err
	}
	return plugin_manage.UninstallPlugin(*dbPlugin)
}

func CheckPluginUpdate(ctx context.Context, id uint) (bool, error) {
	dbPlugin, err := db.GetPluginById(id)
	if err != nil {
		return false, errors.WithMessage(err, "failed get plugin")
	}
	has, found := plugin_manage.CheckUpdate(dbPlugin.UUID, dbPlugin.Version)
	if !found {
		return false, errs.NotFoundPluginByRepository
	}
	return has, nil
}
