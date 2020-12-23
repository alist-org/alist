package test

import (
	"fmt"
	"github.com/Xhofe/alist/alidrive"
	"github.com/Xhofe/alist/bootstrap"
	"github.com/Xhofe/alist/conf"
	"testing"
)

func init() {
	bootstrap.InitLog()
	bootstrap.ReadConf("../conf.yml")
	bootstrap.InitClient()
	bootstrap.InitAliDrive()
}

func TestGetUserInfo(t *testing.T) {
	user,err:= alidrive.GetUserInfo()
	fmt.Println(err)
	fmt.Println(user)
}

func TestGetRoot(t *testing.T) {
	files,err:=alidrive.GetRoot(50,"",conf.OrderUpdatedAt,conf.DESC)
	fmt.Println(err)
	fmt.Println(files)
}

func TestSearch(t *testing.T) {
	files,err:=alidrive.Search("测试文件",50,"")
	fmt.Println(err)
	fmt.Println(files)
}

func TestGet(t *testing.T) {
	file,err:=alidrive.GetFile("5fb7c80e85e4f335cd344008be1b1b5349f74414")
	fmt.Println(err)
	fmt.Println(file)
}