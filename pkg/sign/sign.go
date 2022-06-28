package sign

import "errors"

type Sign interface {
	Sign(data string, expire int64) string
	Verify(data, sign string) error
}

var (
	ErrSignExpired   = errors.New("sign expired")
	ErrSignInvalid   = errors.New("sign invalid")
	ErrExpireInvalid = errors.New("expire invalid")
	ErrExpireMissing = errors.New("expire missing")
)
