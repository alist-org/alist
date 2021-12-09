package base

import (
	"errors"
)

var (
	ErrPathNotFound = errors.New("path not found")
	ErrNotFile      = errors.New("not file")
	ErrNotImplement = errors.New("not implement")
	ErrNotSupport   = errors.New("not support")
	ErrNotFolder    = errors.New("not a folder")
)

const (
	TypeString = "string"
	TypeSelect = "select"
	TypeBool   = "bool"
	TypeNumber = "number"
)

type Json map[string]interface{}

type TokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Header struct{
	Name string `json:"name"`
	Value string `json:"value"`
}

type Link struct {
	Url string `json:"url"`
	Headers []Header `json:"headers"`
}