package driver

import (
	log "github.com/sirupsen/logrus"
)

type New func() Driver

var driversMap = map[string]New{}

func RegisterDriver(config Config, driver New) {
	log.Infof("register driver: [%s]", config.Name)
	driversMap[config.Name] = driver
}
