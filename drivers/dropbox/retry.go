package template

import (
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/auth"
)

type f[A any, R any] func(a A) (R, error)

type retry[A any, R any] struct {
	f           f[A, R]
	beforeRetry func()
}

func (r retry[A, R]) call(arg A) (R, error) {
	res, err := r.f(arg)
	if err == nil || r.beforeRetry == nil || !utils.SliceContains([]string{
		auth.AuthErrorExpiredAccessToken,
		auth.AuthErrorInvalidAccessToken,
	}, err.Error()) {
		return res, nil
	}
	r.beforeRetry()
	return r.f(arg)
}
