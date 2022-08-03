package handles

import (
	"fmt"

	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

func ListDriverItems(c *gin.Context) {
	common.SuccessResp(c, operations.GetDriverItemsMap())
}

func ListDriverNames(c *gin.Context) {
	common.SuccessResp(c, operations.GetDriverNames())
}

func GetDriverItems(c *gin.Context) {
	driverName := c.Query("driver")
	itemsMap := operations.GetDriverItemsMap()
	items, ok := itemsMap[driverName]
	if !ok {
		common.ErrorStrResp(c, fmt.Sprintf("driver [%s] not found", driverName), 404)
		return
	}
	common.SuccessResp(c, items)
}
