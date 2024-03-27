package s3

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type TmpTokenResponse struct {
	Code int                  `json:"code"`
	Msg  string               `json:"msg"`
	Data TmpTokenResponseData `json:"data,omitempty"`
}
type TmpTokenResponseData struct {
	Credentials Credentials `json:"Credentials"`
	ExpiredAt   int         `json:"ExpiredAt"`
}
type Credentials struct {
	AccessKeyId     string `json:"accessKeyId,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	SessionToken    string `json:"sessionToken,omitempty"`
}

func getCredentials(AccessKey, SecretKey string) (rst Credentials, err error) {
	apiPath := "/auth/tmp_token.json"
	reqBody, err := json.Marshal(map[string]interface{}{"channel": "OSS_FULL", "scopes": []string{"*"}})
	if err != nil {
		return rst, err
	}

	signStr := apiPath + "\n" + string(reqBody)
	hmacObj := hmac.New(sha1.New, []byte(SecretKey))
	hmacObj.Write([]byte(signStr))
	sign := hex.EncodeToString(hmacObj.Sum(nil))
	Authorization := "TOKEN " + AccessKey + ":" + sign

	req, err := http.NewRequest("POST", "https://api.dogecloud.com"+apiPath, strings.NewReader(string(reqBody)))
	if err != nil {
		return rst, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", Authorization)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return rst, err
	}
	defer resp.Body.Close()
	ret, err := io.ReadAll(resp.Body)
	if err != nil {
		return rst, err
	}
	var tmpTokenResp TmpTokenResponse
	err = json.Unmarshal(ret, &tmpTokenResp)
	if err != nil {
		return rst, err
	}
	return tmpTokenResp.Data.Credentials, nil
}
