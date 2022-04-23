package base

import (
	"errors"
	"io"
	"net/http"
)

var (
	ErrPathNotFound = errors.New("path not found")
	ErrNotFile      = errors.New("not file")
	ErrNotImplement = errors.New("not implement")
	ErrNotSupport   = errors.New("not support")
	ErrNotFolder    = errors.New("not a folder")
	ErrEmptyFile    = errors.New("empty file")
	ErrRelativePath = errors.New("access using relative path is not allowed")
	ErrEmptyToken   = errors.New("empty token")
)

const (
	TypeString = "string"
	TypeSelect = "select"
	TypeBool   = "bool"
	TypeNumber = "number"
	TypeText   = "text"
)

const (
	Get = iota
	Post
	Put
	Delete
	Patch
)

type Json map[string]interface{}

type TokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Link struct {
	Url      string   `json:"url"`
	Headers  []Header `json:"headers"`
	Data     io.ReadCloser
	FilePath string `json:"path"` // for native
	Status   int
	Header   http.Header
}
