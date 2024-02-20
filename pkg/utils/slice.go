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

// SliceAllContains check if slice all contains elements
func SliceAllContains[T comparable](arr []T, vs ...T) bool {
	vsMap := make(map[T]struct{})
	for _, v := range arr {
		vsMap[v] = struct{}{}
	}
	for _, v := range vs {
		if _, ok := vsMap[v]; !ok {
			return false
		}
	}
	return true
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

func SliceMeet[T1, T2 any](arr []T1, v T2, meet func(item T1, v T2) bool) bool {
	for _, item := range arr {
		if meet(item, v) {
			return true
		}
	}
	return false
}

func SliceFilter[T any](arr []T, filter func(src T) bool) []T {
	res := make([]T, 0, len(arr))
	for _, src := range arr {
		if filter(src) {
			res = append(res, src)
		}
	}
	return res
}

func SliceReplace[T any](arr []T, replace func(src T) T) {
	for i, src := range arr {
		arr[i] = replace(src)
	}
}
