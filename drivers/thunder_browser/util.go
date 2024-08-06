package thunder_browser

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

const (
	API_URL        = "https://x-api-pan.xunlei.com/drive/v1"
	FILE_API_URL   = API_URL + "/files"
	XLUSER_API_URL = "https://xluser-ssl.xunlei.com/v1"
)

var Algorithms = []string{
	"uWRwO7gPfdPB/0NfPtfQO+71",
	"F93x+qPluYy6jdgNpq+lwdH1ap6WOM+nfz8/V",
	"0HbpxvpXFsBK5CoTKam",
	"dQhzbhzFRcawnsZqRETT9AuPAJ+wTQso82mRv",
	"SAH98AmLZLRa6DB2u68sGhyiDh15guJpXhBzI",
	"unqfo7Z64Rie9RNHMOB",
	"7yxUdFADp3DOBvXdz0DPuKNVT35wqa5z0DEyEvf",
	"RBG",
	"ThTWPG5eC0UBqlbQ+04nZAptqGCdpv9o55A",
}

const (
	ClientID          = "ZUBzD9J_XPXfn7f7"
	ClientSecret      = "yESVmHecEe6F0aou69vl-g"
	ClientVersion     = "1.10.0.2633"
	PackageName       = "com.xunlei.browser"
	DownloadUserAgent = "AndroidDownloadManager/13 (Linux; U; Android 13; M2004J7AC Build/SP1A.210812.016)"
	SdkVersion        = "233100"
)

const (
	FOLDER    = "drive#folder"
	FILE      = "drive#file"
	RESUMABLE = "drive#resumable"
)

const (
	UPLOAD_TYPE_UNKNOWN = "UPLOAD_TYPE_UNKNOWN"
	//UPLOAD_TYPE_FORM      = "UPLOAD_TYPE_FORM"
	UPLOAD_TYPE_RESUMABLE = "UPLOAD_TYPE_RESUMABLE"
	UPLOAD_TYPE_URL       = "UPLOAD_TYPE_URL"
)

const (
	ThunderDriveSpace                 = ""
	ThunderDriveSafeSpace             = "SPACE_SAFE"
	ThunderBrowserDriveSpace          = "SPACE_BROWSER"
	ThunderBrowserDriveSafeSpace      = "SPACE_BROWSER_SAFE"
	ThunderDriveFolderType            = "DEFAULT_ROOT"
	ThunderBrowserDriveSafeFolderType = "BROWSER_SAFE"
)

func GetAction(method string, url string) string {
	urlpath := regexp.MustCompile(`://[^/]+((/[^/\s?#]+)*)`).FindStringSubmatch(url)[1]
	return method + ":" + urlpath
}

type Common struct {
	client *resty.Client

	captchaToken string

	// 签名相关,二选一
	Algorithms             []string
	Timestamp, CaptchaSign string

	// 必要值,签名相关
	DeviceID          string
	ClientID          string
	ClientSecret      string
	ClientVersion     string
	PackageName       string
	UserAgent         string
	DownloadUserAgent string
	UseVideoUrl       bool
	RemoveWay         string

	// 验证码token刷新成功回调
	refreshCTokenCk func(token string)
}

func (c *Common) SetDeviceID(deviceID string) {
	c.DeviceID = deviceID
}

func (c *Common) SetCaptchaToken(captchaToken string) {
	c.captchaToken = captchaToken
}
func (c *Common) GetCaptchaToken() string {
	return c.captchaToken
}

// RefreshCaptchaTokenAtLogin 刷新验证码token(登录后)
func (c *Common) RefreshCaptchaTokenAtLogin(action, userID string) error {
	metas := map[string]string{
		"client_version": c.ClientVersion,
		"package_name":   c.PackageName,
		"user_id":        userID,
	}
	metas["timestamp"], metas["captcha_sign"] = c.GetCaptchaSign()
	return c.refreshCaptchaToken(action, metas)
}

// RefreshCaptchaTokenInLogin 刷新验证码token(登录时)
func (c *Common) RefreshCaptchaTokenInLogin(action, username string) error {
	metas := make(map[string]string)
	if ok, _ := regexp.MatchString(`\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`, username); ok {
		metas["email"] = username
	} else if len(username) >= 11 && len(username) <= 18 {
		metas["phone_number"] = username
	} else {
		metas["username"] = username
	}
	return c.refreshCaptchaToken(action, metas)
}

// GetCaptchaSign 获取验证码签名
func (c *Common) GetCaptchaSign() (timestamp, sign string) {
	if len(c.Algorithms) == 0 {
		return c.Timestamp, c.CaptchaSign
	}
	timestamp = fmt.Sprint(time.Now().UnixMilli())
	str := fmt.Sprint(c.ClientID, c.ClientVersion, c.PackageName, c.DeviceID, timestamp)
	for _, algorithm := range c.Algorithms {
		str = utils.GetMD5EncodeStr(str + algorithm)
	}
	sign = "1." + str
	return
}

// 刷新验证码token
func (c *Common) refreshCaptchaToken(action string, metas map[string]string) error {
	param := CaptchaTokenRequest{
		Action:       action,
		CaptchaToken: c.captchaToken,
		ClientID:     c.ClientID,
		DeviceID:     c.DeviceID,
		Meta:         metas,
		RedirectUri:  "xlaccsdk01://xunlei.com/callback?state=harbor",
	}
	var e ErrResp
	var resp CaptchaTokenResponse
	_, err := c.Request(XLUSER_API_URL+"/shield/captcha/init", http.MethodPost, func(req *resty.Request) {
		req.SetError(&e).SetBody(param)
	}, &resp)

	if err != nil {
		return err
	}

	if e.IsError() {
		return &e
	}

	if resp.Url != "" {
		return fmt.Errorf(`need verify: <a target="_blank" href="%s">Click Here</a>`, resp.Url)
	}

	if resp.CaptchaToken == "" {
		return fmt.Errorf("empty captchaToken")
	}

	if c.refreshCTokenCk != nil {
		c.refreshCTokenCk(resp.CaptchaToken)
	}
	c.SetCaptchaToken(resp.CaptchaToken)
	return nil
}

// Request 只有基础信息的请求
func (c *Common) Request(url, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := c.client.R().SetHeaders(map[string]string{
		"user-agent":       c.UserAgent,
		"accept":           "application/json;charset=UTF-8",
		"x-device-id":      c.DeviceID,
		"x-client-id":      c.ClientID,
		"x-client-version": c.ClientVersion,
	})

	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}

	var erron ErrResp
	utils.Json.Unmarshal(res.Body(), &erron)
	if erron.IsError() {
		return nil, &erron
	}

	return res.Body(), nil
}

// 计算文件Gcid
func getGcid(r io.Reader, size int64) (string, error) {
	calcBlockSize := func(j int64) int64 {
		var psize int64 = 0x40000
		for float64(j)/float64(psize) > 0x200 && psize < 0x200000 {
			psize = psize << 1
		}
		return psize
	}

	hash1 := sha1.New()
	hash2 := sha1.New()
	readSize := calcBlockSize(size)
	for {
		hash2.Reset()
		if n, err := utils.CopyWithBufferN(hash2, r, readSize); err != nil && n == 0 {
			if err != io.EOF {
				return "", err
			}
			break
		}
		hash1.Write(hash2.Sum(nil))
	}
	return hex.EncodeToString(hash1.Sum(nil)), nil
}

type CustomTime struct {
	time.Time
}

const timeFormat = time.RFC3339

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	str := string(b)
	if str == `""` {
		*ct = CustomTime{Time: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)}
		return nil
	}

	t, err := time.Parse(`"`+timeFormat+`"`, str)
	if err != nil {
		return err
	}
	*ct = CustomTime{Time: t}
	return nil
}

// EncryptPassword 超级保险箱 加密
func EncryptPassword(password string) string {
	if password == "" {
		return ""
	}
	// 将字符串转换为字节数组
	byteData := []byte(password)
	// 计算MD5哈希值
	hash := md5.Sum(byteData)
	// 将哈希值转换为十六进制字符串
	return hex.EncodeToString(hash[:])
}

func generateDeviceSign(deviceID, packageName string) string {

	signatureBase := fmt.Sprintf("%s%s%s%s", deviceID, packageName, "22062", "a5d7416858147a4ab99573872ffccef8")

	sha1Hash := sha1.New()
	sha1Hash.Write([]byte(signatureBase))
	sha1Result := sha1Hash.Sum(nil)

	sha1String := hex.EncodeToString(sha1Result)

	md5Hash := md5.New()
	md5Hash.Write([]byte(sha1String))
	md5Result := md5Hash.Sum(nil)

	md5String := hex.EncodeToString(md5Result)

	deviceSign := fmt.Sprintf("div101.%s%s", deviceID, md5String)

	return deviceSign
}

func BuildCustomUserAgent(deviceID, appName, sdkVersion, clientVersion, packageName string) string {
	//deviceSign := generateDeviceSign(deviceID, packageName)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("ANDROID-%s/%s ", appName, clientVersion))
	sb.WriteString("networkType/WIFI ")
	sb.WriteString(fmt.Sprintf("appid/%s ", "22062"))
	sb.WriteString(fmt.Sprintf("deviceName/Xiaomi_M2004j7ac "))
	sb.WriteString(fmt.Sprintf("deviceModel/M2004J7AC "))
	sb.WriteString(fmt.Sprintf("OSVersion/13 "))
	sb.WriteString(fmt.Sprintf("protocolVersion/301 "))
	sb.WriteString(fmt.Sprintf("platformversion/10 "))
	sb.WriteString(fmt.Sprintf("sdkVersion/%s ", sdkVersion))
	sb.WriteString(fmt.Sprintf("Oauth2Client/0.9 (Linux 4_9_337-perf-sn-uotan-gd9d488809c3d) (JAVA 0) "))
	return sb.String()
}
