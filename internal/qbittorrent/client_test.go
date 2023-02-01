package qbittorrent

import "testing"

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
