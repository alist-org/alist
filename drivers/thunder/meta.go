package thunder

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
)

// 高级设置
type ExpertAddition struct {
	driver.RootID

	LoginType string `json:"login_type" type:"select" options:"user,refresh_token" default:"user"`
	SignType  string `json:"sign_type" type:"select" options:"algorithms,captcha_sign" default:"algorithms"`

	// 登录方式1
	Username string `json:"username" required:"true" help:"login type is user,this is required"`
	Password string `json:"password" required:"true" help:"login type is user,this is required"`
	// 登录方式2
	RefreshToken string `json:"refresh_token" required:"true" help:"login type is refresh_token,this is required"`

	// 签名方法1
	Algorithms string `json:"algorithms" required:"true" help:"sign type is algorithms,this is required" default:"HPxr4BVygTQVtQkIMwQH33ywbgYG5l4JoR,GzhNkZ8pOBsCY+7,v+l0ImTpG7c7/,e5ztohgVXNP,t,EbXUWyVVqQbQX39Mbjn2geok3/0WEkAVxeqhtx857++kjJiRheP8l77gO,o7dvYgbRMOpHXxCs,6MW8TD8DphmakaxCqVrfv7NReRRN7ck3KLnXBculD58MvxjFRqT+,kmo0HxCKVfmxoZswLB4bVA/dwqbVAYghSb,j,4scKJNdd7F27Hv7tbt"`
	// 签名方法2
	CaptchaSign string `json:"captcha_sign" required:"true" help:"sign type is captcha_sign,this is required"`
	Timestamp   string `json:"timestamp" required:"true" help:"sign type is captcha_sign,this is required"`

	// 验证码
	CaptchaToken string `json:"captcha_token"`

	// 必要且影响登录,由签名决定
	DeviceID      string `json:"device_id"  required:"true" default:"9aa5c268e7bcfc197a9ad88e2fb330e5"`
	ClientID      string `json:"client_id"  required:"true" default:"Xp6vsxz_7IYVw2BB"`
	ClientSecret  string `json:"client_secret"  required:"true" default:"Xp6vsy4tN9toTVdMSpomVdXpRmES"`
	ClientVersion string `json:"client_version"  required:"true" default:"7.51.0.8196"`
	PackageName   string `json:"package_name"  required:"true" default:"com.xunlei.downloadprovider"`

	//不影响登录,影响下载速度
	UserAgent         string `json:"user_agent"  required:"true" default:"ANDROID-com.xunlei.downloadprovider/7.51.0.8196 netWorkType/4G appid/40 deviceName/Xiaomi_M2004j7ac deviceModel/M2004J7AC OSVersion/12 protocolVersion/301 platformVersion/10 sdkVersion/220200 Oauth2Client/0.9 (Linux 4_14_186-perf-gdcf98eab238b) (JAVA 0)"`
	DownloadUserAgent string `json:"download_user_agent"  required:"true" default:"Dalvik/2.1.0 (Linux; U; Android 12; M2004J7AC Build/SP1A.210812.016)"`

	//优先使用视频链接代替下载链接
	UseVideoUrl bool `json:"use_video_url"`
}

// 登录特征,用于判断是否重新登录
func (i *ExpertAddition) GetIdentity() string {
	hash := md5.New()
	if i.LoginType == "refresh_token" {
		hash.Write([]byte(i.RefreshToken))
	} else {
		hash.Write([]byte(i.Username + i.Password))
	}

	if i.SignType == "captcha_sign" {
		hash.Write([]byte(i.CaptchaSign + i.Timestamp))
	} else {
		hash.Write([]byte(i.Algorithms))
	}

	hash.Write([]byte(i.DeviceID))
	hash.Write([]byte(i.ClientID))
	hash.Write([]byte(i.ClientSecret))
	hash.Write([]byte(i.ClientVersion))
	hash.Write([]byte(i.PackageName))
	return hex.EncodeToString(hash.Sum(nil))
}

type Addition struct {
	driver.RootID
	Username     string `json:"username" required:"true"`
	Password     string `json:"password" required:"true"`
	CaptchaToken string `json:"captcha_token"`
}

// 登录特征,用于判断是否重新登录
func (i *Addition) GetIdentity() string {
	return utils.GetMD5EncodeStr(i.Username + i.Password)
}

var config = driver.Config{
	Name:      "Thunder",
	LocalSort: true,
	OnlyProxy: true,
}

var configExpert = driver.Config{
	Name:      "ThunderExpert",
	LocalSort: true,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Thunder{}
	})
	op.RegisterDriver(func() driver.Driver {
		return &ThunderExpert{}
	})
}
