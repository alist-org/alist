package main

import (
	"fmt"
	"net/url"
	"testing"
)

func TestUrl(t *testing.T) {
	s,_ := url.QueryUnescape("/ali/%E7%8C%AA%E5%A4%B4%E7%9A%84%E6%96%87%E4%BB%B6%5B%E5%98%BF%E5%98%BF%5D/%E9%82%B9%E9%82%B9%E7%9A%84%E6%96%87%E4%BB%B6/%E6%A1%8C%E9%9D%A2%E5%A3%81%E7%BA%B8/v2-e8f266ba17ae387eefed1cb22b2b5e4e_r.jpg")
	fmt.Print(s)
}