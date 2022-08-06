package utils

import (
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
)

func GetSignInUrl(redirectUri string) string {
	return casdoorsdk.GetSigninUrl(redirectUri)
}
