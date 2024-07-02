// Credits: https://pkg.go.dev/github.com/rclone/rclone@v1.65.2/cmd/serve/s3
// Package s3 implements a fake s3 server for alist
package s3

import (
	"context"
	"math/rand"
	"net/http"

	"github.com/alist-org/gofakes3"
)

// Make a new S3 Server to serve the remote
func NewServer(ctx context.Context) (h http.Handler, err error) {
	var newLogger logger
	faker := gofakes3.New(
		newBackend(),
		// gofakes3.WithHostBucket(!opt.pathBucketMode),
		gofakes3.WithLogger(newLogger),
		gofakes3.WithRequestID(rand.Uint64()),
		gofakes3.WithoutVersioning(),
		gofakes3.WithV4Auth(authlistResolver()),
		gofakes3.WithIntegrityCheck(true), // Check Content-MD5 if supplied
	)

	return faker.Server(), nil
}
