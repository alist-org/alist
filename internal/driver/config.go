package driver

type Config struct {
	Name              string `json:"name"`
	LocalSort         bool   `json:"local_sort"`
	OnlyLocal         bool   `json:"only_local"`
	OnlyProxy         bool   `json:"only_proxy"`
	NoCache           bool   `json:"no_cache"`
	NoUpload          bool   `json:"no_upload"`
	NeedMs            bool   `json:"need_ms"` // if need get message from user, such as validate code
	DefaultRoot       string `json:"default_root"`
	CheckStatus       bool   `json:"-"`
	Alert             string `json:"alert"` //info,success,warning,danger
	NoOverwriteUpload bool   `json:"-"`     // whether to support overwrite upload
	ProxyRangeOption  bool   `json:"-"`
}

func (c Config) MustProxy() bool {
	return c.OnlyProxy || c.OnlyLocal
}
