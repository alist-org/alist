package op

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type SendNotifyPlatform struct{}

// 注意映射方法名必需大写要不然找不到
func (e SendNotifyPlatform) Bark(body string, title string, content string) (bool, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(body), &m)
	barkPush, barkIcon, barkSound, barkGroup, barkLevel, barkUrl := m["barkPush"].(string), m["barkIcon"].(string), m["barkSound"].(string), m["barkGroup"].(string), m["barkLevel"].(string), m["barkUrl"].(string)

	if !strings.HasPrefix(barkPush, "http") {
		barkPush = fmt.Sprintf("https://api.day.app/%s", barkPush)
	}

	urlValues := url.Values{}
	urlValues.Set("icon", barkIcon)
	urlValues.Set("sound", barkSound)
	urlValues.Set("group", barkGroup)
	urlValues.Set("level", barkLevel)
	urlValues.Set("url", barkUrl)
	url := fmt.Sprintf("%s/%s/%s?%s", barkPush, url.QueryEscape(title), url.QueryEscape(content), urlValues.Encode())

	resp, err := http.Get(url)
	if err != nil {
		log.Error("通知发送失败")
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
}

// 注意映射方法名必需大写要不然找不到
func (e SendNotifyPlatform) Gotify(body string, title string, content string) (bool, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(body), &m)
	gotifyUrl, gotifyToken, gotifyPriority := m["gotifyUrl"].(string), m["gotifyToken"].(string), m["gotifyPriority"].(string)

	surl := fmt.Sprintf("%s/message?token=%s", gotifyUrl, gotifyToken)
	data := url.Values{}
	data.Set("title", title)
	data.Set("message", content)
	data.Set("priority", fmt.Sprintf("%d", gotifyPriority))

	req, err := http.NewRequest("POST", surl, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if err != nil {
		log.Error("通知发送失败")
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
}

// 注意映射方法名必需大写要不然找不到
func (e SendNotifyPlatform) GoCqHttpBot(body string, title string, content string) (bool, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(body), &m)
	goCqHttpBotUrl, goCqHttpBotToken, goCqHttpBotQq := m["goCqHttpBotUrl"].(string), m["goCqHttpBotToken"].(string), m["goCqHttpBotQq"].(string)

	surl := fmt.Sprintf("%s?user_id=%s", goCqHttpBotUrl, goCqHttpBotQq)
	data := map[string]string{"message": fmt.Sprintf("%s\n%s", title, content)}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", surl, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+goCqHttpBotToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if err != nil {
		log.Error("通知发送失败")
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
}

// 注意映射方法名必需大写要不然找不到
func (e SendNotifyPlatform) ServerChan(body string, title string, content string) (bool, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(body), &m)
	serverChanKey := m["serverChanKey"].(string)

	surl := ""
	if len(serverChanKey) >= 3 && serverChanKey[:3] == "SCT" {
		surl = fmt.Sprintf("https://sctapi.ftqq.com/%s.send", serverChanKey)
	} else {
		surl = fmt.Sprintf("https://sc.ftqq.com/%s.send", serverChanKey)
	}

	data := url.Values{}
	data.Set("title", title)
	data.Set("desp", content)

	req, err := http.NewRequest("POST", surl, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if err != nil {
		log.Error("通知发送失败")
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
}

// 注意映射方法名必需大写要不然找不到
func (e SendNotifyPlatform) PushDeer(body string, title string, content string) (bool, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(body), &m)
	pushDeerKey, pushDeerUrl := m["pushDeerKey"].(string), m["pushDeerUrl"].(string)

	surl := pushDeerUrl
	if surl == "" {
		surl = "https://api2.pushdeer.com/message/push"
	}

	data := url.Values{}
	data.Set("pushkey", pushDeerKey)
	data.Set("text", title)
	data.Set("desp", content)
	data.Set("type", "markdown")

	req, err := http.NewRequest("POST", surl, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if err != nil {
		log.Error("通知发送失败")
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
}

// 注意映射方法名必需大写要不然找不到
func (e SendNotifyPlatform) TelegramBot(body string, title string, content string) (bool, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(body), &m)
	telegramBotToken, telegramBotUserId, telegramBotProxyHost, telegramBotProxyPort, telegramBotProxyAuth, telegramBotApiHost := m["telegramBotToken"].(string), m["telegramBotUserId"].(string), m["telegramBotProxyHost"].(string), m["telegramBotProxyPort"].(string), m["telegramBotProxyAuth"].(string), m["telegramBotApiHost"].(string)

	if telegramBotApiHost == "" {
		telegramBotApiHost = "https://api.telegram.org"
	}

	surl := fmt.Sprintf("%s/bot%s/sendMessage", telegramBotApiHost, telegramBotToken)

	var client *http.Client
	if telegramBotProxyHost != "" && telegramBotProxyPort != "" {
		proxyURL := fmt.Sprintf("http://%s:%s", telegramBotProxyHost, telegramBotProxyPort)
		if telegramBotProxyAuth != "" {
			proxyURL = fmt.Sprintf("http://%s@%s:%s", telegramBotProxyAuth, telegramBotProxyHost, telegramBotProxyPort)
		}

		proxy := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(proxyURL)
		}

		client = &http.Client{
			Transport: &http.Transport{
				Proxy: proxy,
			},
		}
	} else {
		client = http.DefaultClient
	}

	data := url.Values{}
	data.Set("chat_id", telegramBotUserId)
	data.Set("text", fmt.Sprintf("%s\n\n%s", title, content))
	data.Set("disable_web_page_preview", "true")

	req, err := http.NewRequest("POST", surl, strings.NewReader(data.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if err != nil {
		log.Error("通知发送失败")
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
}

// 注意映射方法名必需大写要不然找不到
func (e SendNotifyPlatform) WeWorkBot(body string, title string, content string) (bool, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(body), &m)
	weWorkBotKey, weWorkOrigin := m["weWorkBotKey"].(string), m["weWorkOrigin"].(string)

	if weWorkOrigin == "" {
		weWorkOrigin = "https://qyapi.weixin.qq.com"
	}

	surl := fmt.Sprintf("%s/cgi-bin/webhook/send?key=%s", weWorkOrigin, weWorkBotKey)

	bodyData := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("%s\n\n%s", title, content),
		},
	}
	data, err := json.Marshal(bodyData)
	if err != nil {
		return false, err
	}

	var client *http.Client
	client = http.DefaultClient

	req, err := http.NewRequest("POST", surl, strings.NewReader(string(data)))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if err != nil {
		log.Error("通知发送失败")
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return false, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
}

var platform model.SettingItem

func Notify(title string, content string) {
	platform, err := GetSettingItemByKey(conf.NotifyPlatform)
	notifyBody, err := GetSettingItemByKey(conf.NotifyValue)
	caser := cases.Title(language.English)
	methodName := caser.String(platform.Value)
	if err != nil {
		log.Error("无法找到配置信息")
	}
	//注意映射方法名必需大写要不然找不到
	// 使用反射获取结构体实例的值
	v := reflect.ValueOf(SendNotifyPlatform{})
	// 检查是否成功获取结构体实例的值
	if v.IsValid() {
		log.Debug("成功获取结构体实例的值")
	} else {
		log.Debug("未能获取结构体实例的值")
		return
	}

	method := v.MethodByName(methodName)
	// 检查方法是否存在
	if !method.IsValid() {
		log.Debug("Method %s not found\n", methodName)
		return
	}
	args := []reflect.Value{reflect.ValueOf(notifyBody.Value), reflect.ValueOf(title), reflect.ValueOf(content)}
	// 调用方法
	method.Call(args)
}
