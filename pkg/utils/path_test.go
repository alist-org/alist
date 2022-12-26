package utils

import "testing"

func TestEncodePath(t *testing.T) {
	t.Log(EncodePath("http://localhost:5244/d/123#.png"))
}

func TestFixAndCleanPath(t *testing.T) {
	datas := map[string]string{
		"":                          "/",
		".././":                     "/",
		"../../.../":                "/...",
		"x//\\y/":                   "/x/y",
		".././.x/.y/.//..x../..y..": "/.x/.y/..x../..y..",
	}
	for key, value := range datas {
		if FixAndCleanPath(key) != value {
			t.Logf("raw %s fix fail", key)
		}
	}
}
