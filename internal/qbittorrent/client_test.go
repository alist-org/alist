package qbittorrent

import (
	"github.com/google/uuid"
	"testing"
)

func TestLogin(t *testing.T) {
	// test logging in with wrong password
	c, err := New("http://admin:admin@127.0.0.1:8080/")
	if err != nil {
		t.Error(err)
	}
	err = c.login()
	if err == nil {
		t.Error(err)
	}

	// test logging in with correct password
	c, err = New("http://admin:adminadmin@127.0.0.1:8080/")
	if err != nil {
		t.Error(err)
	}
	err = c.login()
	if err != nil {
		t.Error(err)
	}
}

// in this test, the `Bypass authentication for clients on localhost` option in qBittorrent webui should be disabled
func TestAuthorized(t *testing.T) {
	// init client
	c, err := New("http://admin:adminadmin@127.0.0.1:8080/")
	if err != nil {
		t.Error(err)
	}

	// test without logging in, which should be unauthorized
	authorized, err := c.authorized()
	if err != nil {
		t.Error(err)
	}
	if authorized {
		t.Error("Should not be authorized")
	}

	// test after logging in
	err = c.login()
	if err != nil {
		t.Error(err)
	}
	authorized, err = c.authorized()
	if err != nil {
		t.Error(err)
	}
	if !authorized {
		t.Error("Should be authorized")
	}
}

func TestAdd(t *testing.T) {
	// init client
	c, err := New("http://admin:adminadmin@127.0.0.1:8080/")
	if err != nil {
		t.Error(err)
	}

	// test add
	err = c.login()
	if err != nil {
		t.Error(err)
	}
	err = c.AddFromLink(
		"https://releases.ubuntu.com/22.04/ubuntu-22.04.1-desktop-amd64.iso.torrent",
		"D:\\qBittorrentDownload\\alist",
		uuid.NewString(),
	)
	if err != nil {
		t.Error(err)
	}
	err = c.AddFromLink(
		"magnet:?xt=urn:btih:375ae3280cd80a8e9d7212e11dfaf7c45069dd35&dn=archlinux-2023.02.01-x86_64.iso",
		"D:\\qBittorrentDownload\\alist",
		uuid.NewString(),
	)
	if err != nil {
		t.Error(err)
	}
}
