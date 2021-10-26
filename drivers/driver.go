package drivers

import "github.com/Xhofe/alist/model"

type Driver interface {
	Path(path string, account *model.Account) (*model.File, []*model.File, error)
	Link(path string, account *model.Account) (string,error)
	Save(account model.Account)
}

var driversMap = map[string]Driver{}

func RegisterDriver(name string, driver Driver) {
	driversMap[name] = driver
}

func GetDriver(name string) (driver Driver, ok bool) {
	driver, ok = driversMap[name]
	return
}

func GetDriverNames() []string {
	names := make([]string, 0)
	for k, _ := range driversMap {
		names = append(names, k)
	}
	return names
}
