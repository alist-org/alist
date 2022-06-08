package driver

import (
	log "github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

type New func() Driver

var driversMap = map[string]New{}
var driverItemsMap = map[string]Items{}

func RegisterDriver(config Config, driver New) {
	log.Infof("register driver: [%s]", config.Name)
	registerDriverItems(config, driver().GetAddition())
	driversMap[config.Name] = driver
}

func registerDriverItems(config Config, addition Additional) {
	tAddition := reflect.TypeOf(addition)
	mainItems := getMainItems(config)
	additionalItems := getAdditionalItems(tAddition)
	driverItemsMap[config.Name] = Items{mainItems, additionalItems}
}

func getMainItems(config Config) []Item {
	items := []Item{{
		Name:     "virtual_path",
		Type:     "string",
		Required: true,
		Help:     "",
	}, {
		Name: "index",
		Type: "int",
		Help: "use to sort",
	}, {
		Name: "down_proxy_url",
		Type: "text",
	}, {
		Name: "webdav_direct",
		Type: "bool",
		Help: "Transfer the WebDAV of this account through the native without redirect",
	}}
	if !config.OnlyProxy && !config.OnlyLocal {
		items = append(items, []Item{{
			Name: "web_proxy",
			Type: "bool",
		}, {
			Name: "webdav_proxy",
			Type: "bool",
		},
		}...)
	}
	if config.LocalSort {
		items = append(items, []Item{{
			Name:   "order_by",
			Type:   "select",
			Values: "name,size,updated_at",
		}, {
			Name:   "order_direction",
			Type:   "select",
			Values: "ASC,DESC",
		}}...)
	}
	items = append(items, Item{
		Name:   "extract_folder",
		Values: "front,back",
		Type:   "select",
	})
	return items
}

func getAdditionalItems(t reflect.Type) []Item {
	var items []Item
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag
		ignore, ok := tag.Lookup("ignore")
		if !ok || ignore == "false" {
			continue
		}
		item := Item{
			Name:     tag.Get("json"),
			Type:     strings.ToLower(field.Name),
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
