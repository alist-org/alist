package driver

import (
	log "github.com/sirupsen/logrus"
)

type New func() Driver

var driversMap = map[string]New{}

func RegisterDriver(name string, new New) {
	log.Infof("register driver: [%s]", name)
	driversMap[name] = new
}
