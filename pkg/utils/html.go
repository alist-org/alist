package utils

import "github.com/microcosm-cc/bluemonday"

var _htmlSanitizePolicy = bluemonday.StrictPolicy()

func SanitizeHTML(s string) string {
	return _htmlSanitizePolicy.Sanitize(s)
}
