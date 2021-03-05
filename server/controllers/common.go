package controllers

import "github.com/gin-gonic/gin"

// common meta response
func MetaResponse(code int, msg string) gin.H {
	return gin.H{
		"meta": gin.H{
			"code": code,
			"msg":  msg,
		},
	}
}

// common data response
func DataResponse(data interface{}) gin.H {
	return gin.H{
		"meta": gin.H{
			"code": 200,
			"msg":  "success",
		},
		"data": data,
	}
}
