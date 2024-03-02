// Credits: https://pkg.go.dev/github.com/rclone/rclone@v1.65.2/cmd/serve/s3
// Package s3 implements a fake s3 server for alist
package s3

import (
	"fmt"

	"github.com/Mikubill/gofakes3"
	"github.com/alist-org/alist/v3/pkg/utils"
)

// logger output formatted message
type logger struct{}

// print log message
func (l logger) Print(level gofakes3.LogLevel, v ...interface{}) {
	switch level {
	default:
		fallthrough
	case gofakes3.LogErr:
		utils.Log.Errorf("serve s3: %s", fmt.Sprintln(v...))
	case gofakes3.LogWarn:
		utils.Log.Infof("serve s3: %s", fmt.Sprintln(v...))
	case gofakes3.LogInfo:
		utils.Log.Debugf("serve s3: %s", fmt.Sprintln(v...))
	}
}
