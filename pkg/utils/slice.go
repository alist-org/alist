package utils

import (
	"strings"

	"github.com/pkg/errors"
)

// SliceEqual check if two slices are equal
func SliceEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// SliceContains check if slice contains element
func SliceContains[T comparable](arr []T, v T) bool {
	for _, vv := range arr {
		if vv == v {
			return true
		}
	}
	return false
}

// SliceConvert convert slice to another type slice
func SliceConvert[S any, D any](srcS []S, convert func(src S) (D, error)) ([]D, error) {
	res := make([]D, 0, len(srcS))
	for i := range srcS {
		dst, err := convert(srcS[i])
		if err != nil {
			return nil, err
		}
		res = append(res, dst)
	}
	return res, nil
}

func MustSliceConvert[S any, D any](srcS []S, convert func(src S) D) []D {
	res := make([]D, 0, len(srcS))
	for i := range srcS {
		dst := convert(srcS[i])
		res = append(res, dst)
	}
	return res
}

func MergeErrors(errs ...error) error {
	errStr := strings.Join(MustSliceConvert(errs, func(err error) string {
		return err.Error()
	}), "\n")
	if errStr != "" {
		return errors.New(errStr)
	}
	return nil
}
