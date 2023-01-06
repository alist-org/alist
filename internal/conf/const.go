package conf

const (
	TypeString = "string"
	TypeSelect = "select"
	TypeBool   = "bool"
	TypeText   = "text"
	TypeNumber = "number"
)

const (
	// site
	VERSION      = "version"
	ApiUrl       = "api_url"
	BasePath     = "base_path"
	SiteTitle    = "site_title"
	Announcement = "announcement"
	AllowIndexed = "allow_indexed"

	Logo      = "logo"
	Favicon   = "favicon"
	MainColor = "main_color"

	// preview
	TextTypes          = "text_types"
	AudioTypes         = "audio_types"
	VideoTypes         = "video_types"
	ImageTypes         = "image_types"
	ProxyTypes         = "proxy_types"
	ProxyIgnoreHeaders = "proxy_ignore_headers"
	AudioAutoplay      = "audio_autoplay"
	VideoAutoplay      = "video_autoplay"

	// global
	HideFiles           = "hide_files"
	CustomizeHead       = "customize_head"
	CustomizeBody       = "customize_body"
	LinkExpiration      = "link_expiration"
	SignAll             = "sign_all"
	PrivacyRegs         = "privacy_regs"
	OcrApi              = "ocr_api"
	FilenameCharMapping = "filename_char_mapping"

	// index
	SearchIndex     = "search_index"
	AutoUpdateIndex = "auto_update_index"
	IndexPaths      = "index_paths"
	IgnorePaths     = "ignore_paths"

	// aria2
	Aria2Uri    = "aria2_uri"
	Aria2Secret = "aria2_secret"

	// single
	Token         = "token"
	IndexProgress = "index_progress"

	//Github
	GithubClientId      = "github_client_id"
	GithubClientSecrets = "github_client_secrets"
	GithubLoginEnabled  = "github_login_enabled"
)

const (
	UNKNOWN = iota
	FOLDER
	//OFFICE
	VIDEO
	AUDIO
	TEXT
	IMAGE
)
