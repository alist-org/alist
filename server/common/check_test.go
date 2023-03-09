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
	}
	for i, data := range datas {
		if IsApply(data.metaPath, data.reqPath, data.applySub) != data.result {
			t.Errorf("TestIsApply %d failed", i)
		}
	}
}
