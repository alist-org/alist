package qbittorrent

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"
)

func TestLogin(t *testing.T) {
	// test logging in with wrong password
	u, err := url.Parse("http://admin:admin@127.0.0.1:8080/")
	if err != nil {
		t.Error(err)
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Error(err)
	}
	var c = &client{
		url:    u,
		client: http.Client{Jar: jar},
	}
	err = c.login()
	if err == nil {
		t.Error(err)
	}

	// test logging in with correct password
	u, err = url.Parse("http://admin:adminadmin@127.0.0.1:8080/")
	if err != nil {
		t.Error(err)
	}
	c.url = u
	err = c.login()
	if err != nil {
		t.Error(err)
	}
}

// in this test, the `Bypass authentication for clients on localhost` option in qBittorrent webui should be disabled
func TestAuthorized(t *testing.T) {
	// init client
	u, err := url.Parse("http://admin:adminadmin@127.0.0.1:8080/")
	if err != nil {
		t.Error(err)
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Error(err)
	}
	var c = &client{
		url:    u,
		client: http.Client{Jar: jar},
	}

	// test without logging in, which should be unauthorized
	authorized := c.authorized()
	if authorized {
		t.Error("Should not be authorized")
	}

	// test after logging in
	err = c.login()
	if err != nil {
		t.Error(err)
	}
	authorized = c.authorized()
	if !authorized {
		t.Error("Should be authorized")
	}
}

func TestNew(t *testing.T) {
	_, err := New("http://admin:adminadmin@127.0.0.1:8080/")
	if err != nil {
		t.Error(err)
	}
	_, err = New("http://admin:wrong_password@127.0.0.1:8080/")
	if err == nil {
		t.Error("Should get an error")
	}
}

func TestAdd(t *testing.T) {
	// init client
	c, err := New("http://admin:adminadmin@127.0.0.1:8080/")
	if err != nil {
		t.Error(err)
	}
	err = c.AddFromLink(
		"https://releases.ubuntu.com/22.04/ubuntu-22.04.1-desktop-amd64.iso.torrent",
		"D:\\qBittorrentDownload\\alist",
		"uuid-1",
	)
	if err != nil {
		t.Error(err)
	}
	err = c.AddFromLink(
		"magnet:?xt=urn:btih:375ae3280cd80a8e9d7212e11dfaf7c45069dd35&dn=archlinux-2023.02.01-x86_64.iso",
		"D:\\qBittorrentDownload\\alist",
		"uuid-2",
	)
	if err != nil {
		t.Error(err)
	}
}

func TestGetInfo(t *testing.T) {
	// init client
	c, err := New("http://admin:adminadmin@127.0.0.1:8080/")
	if err != nil {
		t.Error(err)
	}
	_, err = c.GetInfo("uuid-1")
	if err != nil {
		t.Error(err)
	}
}

func TestGetFiles(t *testing.T) {
	// init client
	c, err := New("http://admin:adminadmin@127.0.0.1:8080/")
	if err != nil {
		t.Error(err)
	}
	files, err := c.GetFiles("uuid-1")
	if err != nil {
		t.Error(err)
	}
	if len(files) != 1 {
		t.Error("should have exactly one file")
	}
}

func TestDelete(t *testing.T) {
	// init client
	c, err := New("http://admin:adminadmin@127.0.0.1:8080/")
	if err != nil {
		t.Error(err)
	}
	err = c.AddFromLink(
		"https://releases.ubuntu.com/22.04/ubuntu-22.04.1-desktop-amd64.iso.torrent",
		"D:\\qBittorrentDownload\\alist",
		"uuid-1",
	)
	if err != nil {
		t.Error(err)
	}
	err = c.Delete("uuid-1", true)
	if err != nil {
		t.Error(err)
	}
}
