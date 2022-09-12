package baiduphoto

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	RefreshToken string `json:"refresh_token" required:"true"`
	ShowType     string `json:"show_type" type:"select" options:"root,root_only_album,root_only_file" default:"root"`
	AlbumID      string `json:"album_id"`
	//AlbumPassword string `json:"album_password"`
	ClientID     string `json:"client_id" required:"true" default:"iYCeC9g08h5vuP9UqvPHKKSVrKFXGa1v"`
	ClientSecret string `json:"client_secret" required:"true" default:"jXiFMOPVPCWlO2M5CwWQzffpNPaGTRBG"`
}

func (a Addition) GetRootId() string {
	return a.AlbumID
}

var config = driver.Config{
	Name:      "BaiduPhoto",
	LocalSort: true,
}

func init() {
	op.RegisterDriver(config, func() driver.Driver {
		return &BaiduPhoto{}
	})
}
