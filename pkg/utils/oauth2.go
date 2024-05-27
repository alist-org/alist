package utils

import "golang.org/x/oauth2"

type tokenSource struct {
	fn func() (*oauth2.Token, error)
}

func (t *tokenSource) Token() (*oauth2.Token, error) {
	return t.fn()
}

func TokenSource(fn func() (*oauth2.Token, error)) oauth2.TokenSource {
	return &tokenSource{fn}
}
