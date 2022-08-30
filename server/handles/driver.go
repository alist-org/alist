package handles

import (
	"fmt"

	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

func ListDriverInfo(c *gin.Context) {
	common.SuccessResp(c, operations.GetDriverInfoMap())
}

func ListDriverNames(c *gin.Context) {
	common.SuccessResp(c, operations.GetDriverNames())
}

func GetDriverInfo(c *gin.Context) {
	driverName := c.Query("driver")
	infoMap := operations.GetDriverInfoMap()
	items, ok := infoMap[driverName]
	if !ok {
		common.ErrorStrResp(c, fmt.Sprintf("driver [%s] not found", driverName), 404)
		return
	}
	common.SuccessResp(c, items)
}
