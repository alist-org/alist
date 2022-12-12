package virtual

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootPath
	NumFile     int   `json:"num_file" type:"number" default:"30" required:"true"`
	NumFolder   int   `json:"num_folder" type:"number" default:"30" required:"true"`
	MaxFileSize int64 `json:"max_file_size" type:"number" default:"1073741824" required:"true"`
	MinFileSize int64 `json:"min_file_size"  type:"number" default:"1048576" required:"true"`
}

var config = driver.Config{
	Name:      "Virtual",
	OnlyLocal: true,
	LocalSort: true,
	NeedMs:    true,
	//NoCache:   true,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Virtual{}
	})
}
