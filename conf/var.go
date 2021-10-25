package conf

import "gorm.io/gorm"

var (
	ConfigFile string // config file
	Conf *Config
	Debug bool

	DB *gorm.DB
)