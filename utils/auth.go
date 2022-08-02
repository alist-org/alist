package utils

import (
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
)

func VerifyAccessToken(accessToken string) bool {
	_, err := casdoorsdk.ParseJwtToken(accessToken)
	return err == nil
}

func GetSignInUrl(redirectUri string) string {
	return casdoorsdk.GetSigninUrl(redirectUri)
}
