package test

import (
	"fmt"
	"strings"
	"testing"
)

func TestSplit(t *testing.T) {
	drive_id:="/123/456"
	strs:=strings.Split(drive_id,"/")
	fmt.Println(strs)
}