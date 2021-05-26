package controllers

import (
	"github.com/Xhofe/alist/alidrive"
)

// common meta response
func MetaResponse(code int, msg string) alidrive.ReqData {
	return alidrive.ReqData{
		Code:    code,
		Data:    nil,
		Message: msg,
	}
}

// common data response
func DataResponse(data interface{})  alidrive.ReqData {
	return alidrive.ReqData{
		Code:    200,
		Data:    data,
		Message: "ok",
	}
}
