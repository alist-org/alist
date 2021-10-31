package drivers

import (
	"encoding/json"
	"github.com/Xhofe/alist/model"
	"github.com/gofiber/fiber/v2"
)

type Driver interface {
	Items() []Item
	Path(path string, account *model.Account) (*model.File, []*model.File, error)
	Link(path string, account *model.Account) (string, error)
	Save(account *model.Account, old *model.Account) error
	Proxy(ctx *fiber.Ctx)
	Preview(path string, account *model.Account) (interface{},error)
	// TODO
	//MakeDir(path string, account *model.Account) error
	//Move(src string, des string, account *model.Account) error
	//Delete(path string) error
	//Upload(file *fs.File, path string, account *model.Account) error
}

type Item struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

var driversMap = map[string]Driver{}

func RegisterDriver(name string, driver Driver) {
	driversMap[name] = driver
}

func GetDriver(name string) (driver Driver, ok bool) {
	driver, ok = driversMap[name]
	return
}

func GetDrivers() map[string][]Item {
	res := make(map[string][]Item, 0)
	for k, v := range driversMap {
		res[k] = v.Items()
	}
	return res
}

type Json map[string]interface{}

func JsonStr(j Json) string {
	data, _ := json.Marshal(j)
	return string(data)
}
