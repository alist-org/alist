package trainbit

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

func get(url string, AUSHELLPORTAL string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(&http.Cookie{
		Name: ".AUSHELLPORTAL",
		Value: AUSHELLPORTAL,
		MaxAge: 2 * 60,
	})
	res, err := http.DefaultClient.Do(req)
	return res, err
}

func postForm(endpoint string, data url.Values, apiKey string, AUSHELLPORTAL string) (*http.Response, error) {
	extData := make(url.Values)
	for key, value := range data {
		extData[key] = make([]string, len(value))
		copy(extData[key], value)
	}
	extData.Set("apikey", apiKey)
	extData.Set("expiredate", "0001-01-06")
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(extData.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name: ".AUSHELLPORTAL",
		Value: AUSHELLPORTAL,
		MaxAge: 2 * 60,
	})
	res, err := http.DefaultClient.Do(req)
	return res, err
}

func getGuid(AUSHELLPORTAL string) (string, error) {
	res, err := get("https://trainbit.com/files/", AUSHELLPORTAL)
	if err != nil {
		return "", err
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	text := string(data)
	reg := regexp.MustCompile(`app.vars.upload.guid = '([0-9a-f]{8}-(?:[0-9a-f]{4}-){3}[0-9a-f]{12})';`)
	result := reg.FindAllStringSubmatch(text, -1)
	return result[0][1], nil
}

func parseRawFileObject(rawObject []any) ([]model.Obj, error) {
	objectList := make([]model.Obj, 0)
	for _, each := range rawObject {
		object := each.(map[string]any)
		if object["id"].(string) == "0" {
			continue
		}
		name, err := base64.URLEncoding.DecodeString(object["name"].(string))
		if err != nil{
			return nil, err
		}
		modified, err := time.Parse("2006/01/02 15:04:05", object["modified"].(string))
		if err != nil{
			return nil, err
		}
		objectList = append(objectList, model.Obj(&model.Object{
			ID: strings.Join([]string{object["id"].(string), strings.Split(object["uploadurl"].(string), "=")[1]}, "_"),
			Name: string(name),
			Size: int64(object["byte"].(float64)),
			Modified: modified,
			IsFolder: int64(object["ty"].(float64)) == 1,
		}))
	}
	return objectList, nil
}