package server

import "github.com/gin-gonic/gin"

func metaResponse(code int, msg string) gin.H {
	return gin.H{
		"meta":gin.H{
			"code":code,
			"msg":msg,
		},
	}
}

func dataResponse(data interface{}) gin.H {
	return gin.H{
		"meta":gin.H{
			"code":200,
			"msg":"success",
		},
		"data":data,
	}
}