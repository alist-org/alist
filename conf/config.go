package conf

// config struct
type Config struct {
	Info struct {
		Title       string `yaml:"title" json:"title"`
		Logo        string `yaml:"logo" json:"logo"`
		FooterText  string `yaml:"footer_text" json:"footer_text"`
		FooterUrl   string `yaml:"footer_url" json:"footer_url"`
		MusicImg    string `yaml:"music_img" json:"music_img"`
		CheckUpdate bool   `yaml:"check_update" json:"check_update"`
		Script      string `yaml:"script" json:"script"`
		Autoplay    bool   `yaml:"autoplay" json:"autoplay"`
		Preview     struct {
			Url        string   `yaml:"url" json:"url"`
			PreProcess []string `yaml:"pre_process" json:"pre_process"`
			Extensions []string `yaml:"extensions" json:"extensions"`
			Text       []string `yaml:"text" json:"text"`
			MaxSize    int      `yaml:"max_size" json:"max_size"`
		} `yaml:"preview" json:"preview"`
	} `yaml:"info"`
	Server struct {
		Port    string `yaml:"port"`                 //端口
		Search  bool   `yaml:"search" json:"search"` //允许搜索
		Static  string `yaml:"static"`
		SiteUrl string `yaml:"site_url" json:"site_url"` //网站url
	} `yaml:"server"`
	AliDrive struct {
		ApiUrl     string `yaml:"api_url"`     //阿里云盘api
		RootFolder string `yaml:"root_folder"` //根目录id
		//Authorization	string		`yaml:"authorization"`//授权token
		LoginToken    string `yaml:"login_token"`
		AccessToken   string `yaml:"access_token"`
		RefreshToken  string `yaml:"refresh_token"`
		MaxFilesCount int    `yaml:"max_files_count"`
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
