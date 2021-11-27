package drivers

import "fmt"

var (
	PathNotFound = fmt.Errorf("path not found")
	NotFile = fmt.Errorf("not file")
)
