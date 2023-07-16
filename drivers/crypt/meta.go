package crypt

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Usually one of two
	//driver.RootPath
	//driver.RootID
	// define other

	FileNameEnc string `json:"filename_encryption" type:"select" required:"true" options:"off,standard,obfuscate" default:"off"`
	DirNameEnc  string `json:"directory_name_encryption" type:"select" required:"true" options:"false,true" default:"false"`
	RemotePath  string `json:"remotePath" required:"true" help:"This is where the encrypted data stores"`

	Password        string `json:"password" required:"true" confidential:"true" help:"same password as Rclone Crypt"`
	Salt            string `json:"salt" confidential:"true"  help:"Password or pass phrase for salt. Optional but recommended"`
	EncryptedSuffix string `json:"encryptedSuffix" required:"true" default:".bin" help:"encrypted files will have this suffix"`
}

/*// inMemory contains decrypted confidential info and other temp data. will not persist these info anywhere
type inMemory struct {
	password string
	salt     string
}*/

var config = driver.Config{
	Name:              "RcloneCrypt",
	LocalSort:         true,
	OnlyLocal:         false,
	OnlyProxy:         true,
	NoCache:           true,
	NoUpload:          false,
	NeedMs:            false,
	DefaultRoot:       "/",
	CheckStatus:       false,
	Alert:             "",
	NoOverwriteUpload: false,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Crypt{}
	})
}
