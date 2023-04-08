package mfs_ipfs

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Usually one of two
	driver.RootPath
	// driver.RootID
	// define other
	// Field string `json:"field" type:"select" required:"true" options:"a,b,c" default:"a"`
	CID      string `json:"cid"`
	PinID    string `json:"pinid"`
	JWToken  string `json:"jwtoken"`
	Endpoint string `json:"endpoint"`
	Gateway  string `json:"gateway" default:"https://dweb.link"`
}

var config = driver.Config{
	Name:              "IPFS",
	LocalSort:         true,
	OnlyLocal:         false,
	OnlyProxy:         false,
	NoCache:           false,
	NoUpload:          false,
	NeedMs:            false,
	DefaultRoot:       "/",
	CheckStatus:       false,
	Alert:             "",
	NoOverwriteUpload: false,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &MfsIpfs{}
	})
}
