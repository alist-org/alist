package drivers

import "fmt"

var (
	PathNotFound = fmt.Errorf("path not found")
	NotFile = fmt.Errorf("not file")
	NotImplement = fmt.Errorf("not implement")
	NotSupport = fmt.Errorf("not support")
)

const (
	STRING = "string"
	SELECT = "select"
	BOOL = "bool"
)
