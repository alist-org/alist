package plugin_yaegi

import (
	"fmt"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	plugin_manage "github.com/alist-org/alist/v3/internal/plugin/manage"
	"github.com/alist-org/alist/v3/internal/plugin/yaegi/lib"
	"github.com/pkg/errors"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func LoadYaegiPlugin(path string) (model.PluginControl, error) {
	i := interp.New(interp.Options{GoPath: conf.Conf.PluginDir})
	// 插件需要使用标准库
	if err := i.Use(stdlib.Symbols); err != nil {
		return nil, err
	}

	// alist 库
	if err := i.Use(lib.Symbols); err != nil {
		return nil, err
	}

	// 导出 PluginControl
	_, err := i.Eval(fmt.Sprintf(`import p "%s"`, path))
	if err != nil {
		return nil, err
	}
	load, err := i.Eval("p.Plugin")
	if err != nil {
		return nil, err
	}

	if control, ok := load.Interface().(func() (model.PluginControl, error)); ok {
		return control()
	}
	return nil, errors.New("Unable to export Plugin method: (func() (model.PluginControl, error))")
}

func init() {
	plugin_manage.RegisterPluginControlManage(model.PLUGIN_MODE_YAEGI, LoadYaegiPlugin)
}
