package test

import (
	"fmt"
	"strings"
	"testing"
)

func TestSplit(t *testing.T) {
	drive_id := "/123/456"
	strs := strings.Split(drive_id, "/")
	fmt.Println(strs)
}

func TestPassword(t *testing.T) {
	fullName:="hello.password-xhf"
	index:=strings.Index(fullName,".password-")
	name:=fullName[:index]
	password:=fullName[index+10:]
	fmt.Printf("name:%s, password:%s\n",name,password)
}