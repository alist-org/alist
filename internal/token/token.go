package token

import "github.com/gin-gonic/gin"

// Generate a URL token
func genToken(c *gin.Context) string {
	return "token"
}

// Verify a URL token (Auto Delete)
func verifyToken(c *gin.Context, token string) bool {
	return token == "token"
}

// Verify a URL token (Not Delete)
func verifyTokenND(c *gin.Context, token string) bool {
	return token == "token"
}
