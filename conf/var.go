package conf

import (
	"context"
	"github.com/eko/gocache/v2/cache"
	"github.com/fatih/color"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"strconv"
)

var (
	BuiltAt   string
	GoVersion string
	GitAuthor string
	GitCommit string
	GitTag    string = "dev"
	WebTag    string
)

var (
	ConfigFile string // config file
	Conf       *Config
	Debug      bool
	Version    bool
	Password   bool
	Docker     bool

	DB    *gorm.DB
	Cache *cache.Cache
	Ctx   = context.TODO()
	Cron  *cron.Cron

	C = color.New(color.FgHiBlue, color.Bold, color.BgHiWhite, color.Underline)
)

var (
	TextTypes = []string{"txt", "htm", "html", "xml", "java", "properties", "sql",
		"js", "md", "json", "conf", "ini", "vue", "php", "py", "bat", "gitignore", "yml",
		"go", "sh", "c", "cpp", "h", "hpp", "tsx", "vtt", "srt", "ass"}
	DProxyTypes = []string{"m3u8"}
	OfficeTypes = []string{"doc", "docx", "xls", "xlsx", "ppt", "pptx", "pdf"}
	VideoTypes  = []string{"mp4", "mkv", "avi", "mov", "rmvb", "webm", "flv"}
	AudioTypes  = []string{"mp3", "flac", "ogg", "m4a", "wav", "opus"}
	ImageTypes  = []string{"jpg", "tiff", "jpeg", "png", "gif", "bmp", "svg", "ico", "swf", "webp"}
)

var settingsMap = make(map[string]string)

func Set(key string, value string) {
	settingsMap[key] = value
}

func GetStr(key string) string {
	value, ok := settingsMap[key]
	if !ok {
		return ""
	}
	return value
}

func GetBool(key string) bool {
	value, ok := settingsMap[key]
	if !ok {
		return false
	}
	return value == "true"
}

func GetInt(key string, defaultV int) int {
	value, ok := settingsMap[key]
	if !ok {
		return defaultV
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return defaultV
	}
	return v
}

var (
	LoadSettings = []string{
		"check parent folder", "check down link", "WebDAV username", "WebDAV password",
		"Visitor WebDAV username", "Visitor WebDAV password",
		"default page size", "load type",
		"ocr api", "favicon",
	}
)

var (
	RawIndexHtml string
	ManageHtml   string
	IndexHtml    string
	Token        string

	//CheckParent        bool
	//CheckDown          bool
	//DavUsername        string
	//DavPassword        string
	//VisitorDavUsername string
	//VisitorDavPassword string
)
