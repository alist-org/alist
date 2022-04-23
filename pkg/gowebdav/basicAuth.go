package gowebdav

import (
	"encoding/base64"
	"net/http"
)

// BasicAuth structure holds our credentials
type BasicAuth struct {
	user string
	pw   string
}

// Type identifies the BasicAuthenticator
func (b *BasicAuth) Type() string {
	return "BasicAuth"
}

// User holds the BasicAuth username
func (b *BasicAuth) User() string {
	return b.user
}

// Pass holds the BasicAuth password
func (b *BasicAuth) Pass() string {
	return b.pw
}

// Authorize the current request
func (b *BasicAuth) Authorize(req *http.Request, method string, path string) {
	a := b.user + ":" + b.pw
	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(a))
	req.Header.Set("Authorization", auth)
}
