package conf

import (
	"context"
	"github.com/eko/gocache/v2/cache"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

var (
	BuiltAt   string
	GoVersion string
	GitAuthor string
	GitCommit string
	GitTag    string = "dev"
)

var (
	ConfigFile string // config file
	Conf       *Config
	Debug      bool
	Version    bool
	Password   bool

	DB    *gorm.DB
	Cache *cache.Cache
	Ctx   = context.TODO()
	Cron  *cron.Cron
)

var (
	TextTypes = []string{"txt", "htm", "html", "xml", "java", "properties", "sql",
		"js", "md", "json", "conf", "ini", "vue", "php", "py", "bat", "gitignore", "yml",
		"go", "sh", "c", "cpp", "h", "hpp", "tsx", "vtt", "srt", "ass"}
	OfficeTypes = []string{"doc", "docx", "xls", "xlsx", "ppt", "pptx", "pdf"}
	VideoTypes  = []string{"mp4", "mkv", "avi", "mov", "rmvb", "webm"}
	AudioTypes  = []string{"mp3", "flac", "ogg", "m4a", "wav"}
	ImageTypes  = []string{"jpg", "tiff", "jpeg", "png", "gif", "bmp", "svg"}
)

// settings
var (
	RawIndexHtml string
	IndexHtml    string
	CheckParent  bool
	CheckDown    bool

	Token       string
	DavUsername string
	DavPassword string
)
