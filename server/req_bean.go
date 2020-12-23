package server

import "github.com/Xhofe/alist/alidrive"

type ListReq struct {
	Password	string	`json:"password"`
	alidrive.ListReq
}