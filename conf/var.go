package conf

import (
	"context"
	"github.com/eko/gocache/v2/cache"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

var (
	ConfigFile string // config file
	Conf       *Config
	Debug      bool

	DB *gorm.DB
	Cache *cache.Cache
	Ctx = context.TODO()
	Cron *cron.Cron
)

var (
	TextTypes   = []string{"txt", "go", "md"}
	OfficeTypes = []string{"doc", "docx", "xls", "xlsx", "ppt", "pptx", "pdf"}
	VideoTypes  = []string{"mp4", "mkv", "avi", "mov", "rmvb"}
	AudioTypes  = []string{"mp3", "flac","ogg"}
	ImageTypes = []string{"jpg","jpeg","png","gif","bmp","svg"}
)
