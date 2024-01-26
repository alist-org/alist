package chaoxing

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/go-resty/resty/v2"
)

func (d *ChaoXing) requestDownload(pathname string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	u := d.conf.DowloadApi + pathname
	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		"Cookie":  d.Cookie,
		"Accept":  "application/json, text/plain, */*",
		"Referer": d.conf.referer,
	})
	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var e Resp
	req.SetError(&e)
	res, err := req.Execute(method, u)
	if err != nil {
		return nil, err
	}
	return res.Body(), nil
}

func (d *ChaoXing) request(pathname string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	u := d.conf.api + pathname
	if strings.Contains(pathname, "getUploadConfig") {
		u = pathname
	}
	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		"Cookie":  d.Cookie,
		"Accept":  "application/json, text/plain, */*",
		"Referer": d.conf.referer,
	})
	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var e Resp
	req.SetError(&e)
	res, err := req.Execute(method, u)
	if err != nil {
		return nil, err
	}
	return res.Body(), nil
}

func (d *ChaoXing) GetFiles(parent string) ([]File, error) {
	files := make([]File, 0)
	query := map[string]string{
		"bbsid":    d.Addition.Bbsid,
		"folderId": parent,
		"recType":  "1",
	}
	var resp ListFileResp
	_, err := d.request("/pc/resource/getResourceList", http.MethodGet, func(req *resty.Request) {
		req.SetQueryParams(query)
	}, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Result != 1 {
		msg := fmt.Sprintf("error code is:%d", resp.Result)
		return nil, errors.New(msg)
	}
	if len(resp.List) > 0 {
		files = append(files, resp.List...)
	}
	querys := map[string]string{
		"bbsid":    d.Addition.Bbsid,
		"folderId": parent,
		"recType":  "2",
	}
	var resps ListFileResp
	_, err = d.request("/pc/resource/getResourceList", http.MethodGet, func(req *resty.Request) {
		req.SetQueryParams(querys)
	}, &resps)
	if err != nil {
		return nil, err
	}
	for _, file := range resps.List {
		// 手机端超星上传的文件没有fileID字段，但ObjectID与fileID相同，可代替
		if file.Content.FileID == "" {
			file.Content.FileID = file.Content.ObjectID
		}
		files = append(files, file)
	}
	return files, nil
}

func EncryptByAES(message, key string) (string, error) {
	aesKey := []byte(key)
	plainText := []byte(message)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	iv := aesKey[:aes.BlockSize]
	mode := cipher.NewCBCEncrypter(block, iv)
	padding := aes.BlockSize - len(plainText)%aes.BlockSize
	paddedText := append(plainText, byte(padding))
	for i := 0; i < padding-1; i++ {
		paddedText = append(paddedText, byte(padding))
	}
	ciphertext := make([]byte, len(paddedText))
	mode.CryptBlocks(ciphertext, paddedText)
	encrypted := base64.StdEncoding.EncodeToString(ciphertext)
	return encrypted, nil
}

func CookiesToString(cookies []*http.Cookie) string {
	var cookieStr string
	for _, cookie := range cookies {
		cookieStr += cookie.Name + "=" + cookie.Value + "; "
	}
	if len(cookieStr) > 2 {
		cookieStr = cookieStr[:len(cookieStr)-2]
	}
	return cookieStr
}

func (d *ChaoXing) Login() (string, error) {
	transferKey := "u2oh6Vu^HWe4_AES"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	uname, err := EncryptByAES(d.Addition.UserName, transferKey)
	if err != nil {
		return "", err
	}
	password, err := EncryptByAES(d.Addition.Password, transferKey)
	if err != nil {
		return "", err
	}
	err = writer.WriteField("uname", uname)
	if err != nil {
		return "", err
	}
	err = writer.WriteField("password", password)
	if err != nil {
		return "", err
	}
	err = writer.WriteField("t", "true")
	if err != nil {
		return "", err
	}
	err = writer.Close()
	if err != nil {
		return "", err
	}
	// Create the request
	req, err := http.NewRequest("POST", "https://passport2.chaoxing.com/fanyalogin", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Content-Length", fmt.Sprintf("%d", body.Len()))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return CookiesToString(resp.Cookies()), nil

}
