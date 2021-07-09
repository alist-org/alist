package conf

import (
	"gorm.io/gorm"
	"net/http"
)

var (
	Debug      bool   // is debug command
	Help       bool   // is help command
	Version    bool   // is print version command
	ConfigFile string // config file
	SkipUpdate bool   // skip update

	Client *http.Client // request client

	DB *gorm.DB

	Origins []string // allow origins
)

var Conf = new(Config)

const (
	VERSION = "v1.0.6"

	ImageThumbnailProcess = "image/resize,w_50"
	VideoThumbnailProcess = "video/snapshot,t_0,f_jpg,w_50"
	ImageUrlProcess       = "image/resize,w_1920/format,jpeg"
	ASC                   = "ASC"
	DESC                  = "DESC"
	OrderUpdatedAt        = "updated_at"
	OrderCreatedAt        = "created_at"
	OrderSize             = "size"
	OrderName             = "name"
	OrderSearch           = "type ASC,updated_at DESC"
	AccessTokenInvalid    = "AccessTokenInvalid"
	Bearer                = "Bearer\t"
)
