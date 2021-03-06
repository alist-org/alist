package test

import (
	"fmt"
	"github.com/Xhofe/alist/alidrive"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
	"testing"
)

func TestStr(t *testing.T) {
	fmt.Println(".password-"[10:])
}

func TestWriteYml(t *testing.T) {
	alidrive.RefreshToken()
	utils.WriteToYml("../conf.yml", conf.Conf)
}

