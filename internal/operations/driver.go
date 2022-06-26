package operations

import (
	"reflect"
	"strings"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type New func() driver.Driver

var driverNewMap = map[string]New{}
var driverItemsMap = map[string]driver.Items{}

func RegisterDriver(config driver.Config, driver New) {
	log.Infof("register driver: [%s]", config.Name)
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
	for k := range driverItemsMap {
		driverNames = append(driverNames, k)
	}
	return driverNames
}

func GetDriverItemsMap() map[string]driver.Items {
	return driverItemsMap
}

func registerDriverItems(config driver.Config, addition driver.Additional) {
	log.Debugf("addition of %s: %+v", config.Name, addition)
	tAddition := reflect.TypeOf(addition)
	mainItems := getMainItems(config)
	additionalItems := getAdditionalItems(tAddition)
	driverItemsMap[config.Name] = driver.Items{
		Main:       mainItems,
		Additional: additionalItems,
	}
}

func getMainItems(config driver.Config) []driver.Item {
	items := []driver.Item{{
		Name:     "virtual_path",
		Type:     driver.TypeString,
		Required: true,
		Help:     "",
	}, {
		Name: "index",
		Type: driver.TypeNumber,
		Help: "use to sort",
	}, {
		Name: "down_proxy_url",
		Type: driver.TypeText,
	}, {
		Name: "webdav_direct",
		Type: driver.TypeBool,
		Help: "Transfer the WebDAV of this account through the native without redirect",
	}}
	if !config.OnlyProxy && !config.OnlyLocal {
		items = append(items, []driver.Item{{
			Name: "web_proxy",
			Type: driver.TypeBool,
		}, {
			Name: "webdav_proxy",
			Type: driver.TypeBool,
		},
		}...)
	}
	if config.LocalSort {
		items = append(items, []driver.Item{{
			Name:   "order_by",
			Type:   driver.TypeSelect,
			Values: "name,size,modified",
		}, {
			Name:   "order_direction",
			Type:   driver.TypeSelect,
			Values: "ASC,DESC",
		}}...)
	}
	items = append(items, driver.Item{
		Name:   "extract_folder",
		Type:   driver.TypeSelect,
		Values: "front,back",
	})
	return items
}

func getAdditionalItems(t reflect.Type) []driver.Item {
	var items []driver.Item
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Type.Kind() == reflect.Struct {
			items = append(items, getAdditionalItems(field.Type)...)
			continue
		}
		tag := field.Tag
		ignore, ok := tag.Lookup("ignore")
		if ok && ignore == "true" {
			continue
		}
		item := driver.Item{
			Name:     tag.Get("json"),
			Type:     strings.ToLower(field.Type.Name()),
			Default:  tag.Get("default"),
			Values:   tag.Get("values"),
			Required: tag.Get("required") == "true",
			Help:     tag.Get("help"),
		}
		if tag.Get("type") != "" {
			item.Type = tag.Get("type")
		}
		// set default type to string
		if item.Type == "" {
			item.Type = "string"
		}
		items = append(items, item)
	}
	return items
}
