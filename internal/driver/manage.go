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
		Type:     TypeString,
		Required: true,
		Help:     "",
	}, {
		Name: "index",
		Type: TypeNumber,
		Help: "use to sort",
	}, {
		Name: "down_proxy_url",
		Type: TypeText,
	}, {
		Name: "webdav_direct",
		Type: TypeBool,
		Help: "Transfer the WebDAV of this account through the native without redirect",
	}}
	if !config.OnlyProxy && !config.OnlyLocal {
		items = append(items, []Item{{
			Name: "web_proxy",
			Type: TypeBool,
		}, {
			Name: "webdav_proxy",
			Type: TypeBool,
		},
		}...)
	}
	if config.LocalSort {
		items = append(items, []Item{{
			Name:   "order_by",
			Type:   TypeSelect,
			Values: "name,size,modified",
		}, {
			Name:   "order_direction",
			Type:   TypeSelect,
			Values: "ASC,DESC",
		}}...)
	}
	items = append(items, Item{
		Name:   "extract_folder",
		Type:   TypeSelect,
		Values: "front,back",
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
