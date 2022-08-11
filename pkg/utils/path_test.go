package utils

import "testing"

func TestEncodePath(t *testing.T) {
	t.Log(EncodePath("http://localhost:5244/d/123#.png"))
}
