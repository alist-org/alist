package op

import (
	"reflect"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/pkg/errors"
)

type New func() driver.Driver

var driverNewMap = map[string]New{}
var driverInfoMap = map[string]driver.Info{}

func RegisterDriver(config driver.Config, driver New) {
	// log.Infof("register driver: [%s]", config.Name)
	registerDriverItems(config, driver().GetAddition())
	driverNewMap[config.Name] = driver
}

func GetDriverNew(name string) (New, error) {
	n, ok := driverNewMap[name]
	if !ok {
		return nil, errors.Errorf("no driver named: %s", name)
	}
	return n, nil
}

func GetDriverNames() []string {
	var driverNames []string
	for k := range driverInfoMap {
		driverNames = append(driverNames, k)
	}
	return driverNames
}

func GetDriverInfoMap() map[string]driver.Info {
	return driverInfoMap
}

func registerDriverItems(config driver.Config, addition driver.Additional) {
	// log.Debugf("addition of %s: %+v", config.Name, addition)
	tAddition := reflect.TypeOf(addition)
	mainItems := getMainItems(config)
	additionalItems := getAdditionalItems(tAddition, config.DefaultRoot)
	driverInfoMap[config.Name] = driver.Info{
		Common:     mainItems,
		Additional: additionalItems,
		Config:     config,
	}
}

func getMainItems(config driver.Config) []driver.Item {
	items := []driver.Item{{
		Name:     "mount_path",
		Type:     conf.TypeString,
		Required: true,
		Help:     "",
	}, {
		Name: "order",
		Type: conf.TypeNumber,
		Help: "use to sort",
	}, {
		Name: "remark",
		Type: conf.TypeText,
	}}
	if !config.NoCache {
		items = append(items, driver.Item{
			Name:     "cache_expiration",
			Type:     conf.TypeNumber,
			Default:  "30",
			Required: true,
			Help:     "The cache expiration time for this storage",
		})
	}
	if !config.OnlyProxy && !config.OnlyLocal {
		items = append(items, []driver.Item{{
			Name: "web_proxy",
			Type: conf.TypeBool,
		}, {
			Name:     "webdav_policy",
			Type:     conf.TypeSelect,
			Options:  "302_redirect,use_proxy_url,native_proxy",
			Default:  "302_redirect",
			Required: true,
		},
		}...)
	} else {
		items = append(items, driver.Item{
			Name:     "webdav_policy",
			Type:     conf.TypeSelect,
			Default:  "native_proxy",
			Options:  "use_proxy_url,native_proxy",
			Required: true,
		})
	}
	items = append(items, driver.Item{
		Name: "down_proxy_url",
		Type: conf.TypeText,
	})
	if config.LocalSort {
		items = append(items, []driver.Item{{
			Name:    "order_by",
			Type:    conf.TypeSelect,
			Options: "name,size,modified",
		}, {
			Name:    "order_direction",
			Type:    conf.TypeSelect,
			Options: "asc,desc",
		}}...)
	}
	items = append(items, driver.Item{
		Name:    "extract_folder",
		Type:    conf.TypeSelect,
		Options: "front,back",
	})
	return items
}

func getAdditionalItems(t reflect.Type, defaultRoot string) []driver.Item {
	var items []driver.Item
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Type.Kind() == reflect.Struct {
			items = append(items, getAdditionalItems(field.Type, defaultRoot)...)
			continue
		}
		tag := field.Tag
		ignore, ok1 := tag.Lookup("ignore")
		name, ok2 := tag.Lookup("json")
		if (ok1 && ignore == "true") || !ok2 {
			continue
		}
		item := driver.Item{
			Name:     name,
			Type:     strings.ToLower(field.Type.Name()),
			Default:  tag.Get("default"),
			Options:  tag.Get("options"),
			Required: tag.Get("required") == "true",
			Help:     tag.Get("help"),
		}
		if tag.Get("type") != "" {
			item.Type = tag.Get("type")
		}
		if item.Name == "root_folder_id" || item.Name == "root_folder_path" {
			if item.Default == "" {
				item.Default = defaultRoot
			}
			item.Required = item.Default != ""
		}
		// set default type to string
		if item.Type == "" {
			item.Type = "string"
		}
		items = append(items, item)
	}
	return items
}
