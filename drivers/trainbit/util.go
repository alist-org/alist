package trainbit

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

func Get(url string, AUSHELLPORTAL string) (*http.Response, error) {
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

func PostForm(url string, data url.Values, apiKey string, AUSHELLPORTAL string) (*http.Response, error) {
	data.Add("apikey", apiKey)
	data.Add("expiredate", "0001-01-06")
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data.Encode()))
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

func getDownloadLink(id string, AUSHELLPORTAL string) (string, error) {
	res, err := Get(fmt.Sprintf("https://trainbit.com/files/%s/", strings.Split(id, "_")[0]), AUSHELLPORTAL)
	if err != nil {
		return "", err
	}
	return res.Header.Get("Location"), nil
}

func readFolder(id string, AUSHELLPORTAL string, apiKey string) ([]model.Obj, error) {
	form := make(url.Values)
	form.Add("parentid", strings.Split(id, "_")[0])
	res, err := PostForm("https://trainbit.com/lib/api/v1/listoffiles", form, apiKey, AUSHELLPORTAL)
	if err != nil {
		return nil, err
	}
	rawData, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	content := make(map[string]any)
	err = json.Unmarshal(rawData, &content)
	if err != nil {
		return nil, err
	}
	objectList := make([]model.Obj, 0)
	for _, each := range content["items"].([]any) {
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

// func getObjectId(path string, AUSHELLPORTAL string, rootId string) ([]string, error) {
// 	path = filepath.Clean(path)
// 	part := strings.Split(path, "/")
// 	if part[0] == "." {
// 		return []string{rootId}, nil
// 	}
// 	objectId := []string{rootId}
// 	for index, each := range part {
// 		objectList, err := readFolder(objectId[len(objectId) - 1], AUSHELLPORTAL)
// 		if err != nil {
// 			return nil, err
// 		}
// 		flag := false
// 		for _, object := range objectList {
// 			if object.Name == each && (index == len(part) - 1 && !object.IsFolder) {
// 				objectId = append(objectId, object.Id)
// 				flag = true
// 				break
// 			}
// 		}
// 		if !flag {
// 			return nil, errors.New("not found")
// 		}
// 	}
// 	return objectId, nil
// }