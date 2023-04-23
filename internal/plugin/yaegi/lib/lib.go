package lib

import "reflect"

//go:generate yaegi extract github.com/go-resty/resty/v2
//go:generate yaegi extract github.com/sirupsen/logrus
//go:generate yaegi extract github.com/json-iterator/go
//go:generate yaegi extract github.com/pkg/errors

// aws
//go:generate yaegi extract github.com/aws/aws-sdk-go/aws
//go:generate yaegi extract github.com/aws/aws-sdk-go/aws/credentials
//go:generate yaegi extract github.com/aws/aws-sdk-go/aws/session
//go:generate yaegi extract github.com/aws/aws-sdk-go/service/s3/s3manager

//go:generate yaegi extract github.com/alist-org/alist/v3/internal/model
//go:generate yaegi extract github.com/alist-org/alist/v3/internal/driver
//go:generate yaegi extract github.com/alist-org/alist/v3/internal/op
//go:generate yaegi extract github.com/alist-org/alist/v3/internal/errs
//go:generate yaegi extract github.com/alist-org/alist/v3/drivers/base

// utils
//go:generate yaegi extract github.com/alist-org/alist/v3/pkg/utils
//go:generate yaegi extract github.com/alist-org/alist/v3/pkg/utils/random"
//go:generate yaegi extract github.com/alist-org/alist/v3/pkg/cron

// adapter
//go:generate yaegi extract github.com/alist-org/alist/v3/internal/plugin/yaegi/adapter/storage
var Symbols = map[string]map[string]reflect.Value{}
