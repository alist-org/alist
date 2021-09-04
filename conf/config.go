package conf

type Drive struct {
	AccessToken    string `yaml:"-"`
	RefreshToken   string `yaml:"refresh_token"`
	RootFolder     string `yaml:"root_folder"` //根目录id
	Name           string `yaml:"name"`
	Password       string `yaml:"password"`
	Hide           bool   `yaml:"hide"`
	DefaultDriveId string `yaml:"-"`
}

// config struct
type Config struct {
	Info struct {
		Title       string   `yaml:"title" json:"title"`
		Roots       []string `yaml:"-" json:"roots"`
		Logo        string   `yaml:"logo" json:"logo"`
		FooterText  string   `yaml:"footer_text" json:"footer_text"`
		FooterUrl   string   `yaml:"footer_url" json:"footer_url"`
		MusicImg    string   `yaml:"music_img" json:"music_img"`
		CheckUpdate bool     `yaml:"check_update" json:"check_update"`
		Script      string   `yaml:"script" json:"script"`
		Autoplay    bool     `yaml:"autoplay" json:"autoplay"`
		Preview     struct {
			Url        string   `yaml:"url" json:"url"`
			PreProcess []string `yaml:"pre_process" json:"pre_process"`
			Extensions []string `yaml:"extensions" json:"extensions"`
			Text       []string `yaml:"text" json:"text"`
			MaxSize    int      `yaml:"max_size" json:"max_size"`
		} `yaml:"preview" json:"preview"`
	} `yaml:"info"`
	Server struct {
		Address    string `yaml:"address"`
		Port       string `yaml:"port"`     //端口
		Search     bool   `yaml:"search"`   //允许搜索
		Download   bool   `yaml:"download"` //允许下载
		Static     string `yaml:"static"`
		SiteUrl    string `yaml:"site_url"` //网站url
		Password   string `yaml:"password"`
		AllowProxy string `yaml:"allow_proxy"` // 允许代理的后缀
	} `yaml:"server"`
	AliDrive struct {
		ApiUrl        string  `yaml:"api_url"` //阿里云盘api
		MaxFilesCount int     `yaml:"max_files_count"`
		Drives        []Drive `yaml:"drives"`
	} `yaml:"ali_drive"`
	Database struct {
		Type        string `yaml:"type"`
		User        string `yaml:"user"`
		Password    string `yaml:"password"`
		Host        string `yaml:"host"`
		Port        int    `yaml:"port"`
		Name        string `yaml:"name"`
		TablePrefix string `yaml:"tablePrefix"`
		DBFile      string `yaml:"dBFile"`
	} `yaml:"database"`
}
