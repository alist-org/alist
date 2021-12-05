package drivers

import "fmt"

var (
	ErrPathNotFound = fmt.Errorf("path not found")
	ErrNotFile      = fmt.Errorf("not file")
	ErrNotImplement = fmt.Errorf("not implement")
	ErrNotSupport   = fmt.Errorf("not support")
)

const (
	TypeString = "string"
	TypeSelect = "select"
	TypeBool   = "bool"
	TypeNumber = "number"
)
