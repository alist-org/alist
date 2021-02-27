package conf

import (
	"github.com/patrickmn/go-cache"
	"net/http"
)

var(
	Debug bool // is debug command
	Help bool // is help command
	Version bool // is print version command
	Con string // config file
	SkipUpdate bool // skip update

	Client *http.Client // request client
	Authorization string // authorization string

	Cache *cache.Cache // cache

	Origins []string // allow origins
)

var Conf = new(Config)

const (
	VERSION="v0.1.7"

	ImageThumbnailProcess="image/resize,w_50"
	VideoThumbnailProcess="video/snapshot,t_0,f_jpg,w_50"
	ImageUrlProcess="image/resize,w_1920/format,jpeg"
	ASC="ASC"
	DESC="DESC"
	OrderUpdatedAt="updated_at"
	OrderCreatedAt="created_at"
	OrderSize="size"
	OrderName="name"
	OrderSearch="type ASC,updated_at DESC"
	AccessTokenInvalid="AccessTokenInvalid"
	Bearer="Bearer\t"
)