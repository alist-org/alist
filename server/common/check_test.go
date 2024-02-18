package common

import "testing"

func TestIsApply(t *testing.T) {
	datas := []struct {
		metaPath string
		reqPath  string
		applySub bool
		result   bool
	}{
		{
			metaPath: "/",
			reqPath:  "/test",
			applySub: true,
			result:   true,
		},
		{
			metaPath: "/",
			reqPath:  "/test",
			applySub: false,
			result:   true,
		},
		{
			metaPath: "/",
			reqPath:  "/test/aaa",
			applySub: false,
			result:   false,
		},
		{
			metaPath: "/",
			reqPath:  "/test/aaa",
			applySub: true,
			result:   true,
		},
		{
			metaPath: "/",
			reqPath:  "/",
			applySub: false,
			result:   false,
		},
		{
			metaPath: "/",
			reqPath:  "/",
			applySub: true,
			result:   false,
		},
		{
			metaPath: "/test",
			reqPath:  "/test/aaa",
			applySub: false,
			result:   true,
		},
		{
			metaPath: "/test",
			reqPath:  "/test/aaa/bbb",
			applySub: false,
			result:   false,
		},
		{
			metaPath: "/test",
			reqPath:  "/test/aaa/bbb",
			applySub: true,
			result:   true,
		},
		{
			metaPath: "/test",
			reqPath:  "/test",
			applySub: false,
			result:   false,
		},
		{
			metaPath: "/test",
			reqPath:  "/test",
			applySub: true,
			result:   false,
		},
	}
	for i, data := range datas {
		if IsApply(data.metaPath, data.reqPath, data.applySub) != data.result {
			t.Errorf("TestIsApply %d failed", i)
		}
	}
}
