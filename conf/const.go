package conf

import (
	"github.com/patrickmn/go-cache"
	"net/http"
)

var(
	Debug bool
	Help bool
	Con string
	Client *http.Client
	Authorization string

	Cache *cache.Cache

	Origins []string
)

var Conf = new(Config)

const (
	VERSION="v0.1.5"

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