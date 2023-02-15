package utils

import (
	"net/url"
)

func InjectQuery(raw string, query url.Values) (string, error) {
	param := query.Encode()
	if param == "" {
		return raw, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	joiner := "?"
	if u.RawQuery != "" {
		joiner = "&"
	}
	return raw + joiner + param, nil
}
