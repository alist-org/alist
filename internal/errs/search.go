package errs

import "fmt"

var (
	SearchNotAvailable  = fmt.Errorf("search not available")
	BuildIndexIsRunning = fmt.Errorf("build index is running, please try later")
)
