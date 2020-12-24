package conf

type Config struct {
	Info struct{
		Title			string		`yaml:"title" json:"title"`
		SiteUrl 		string 		`yaml:"site_url" json:"site_url"`//网站url
		Logo			string		`yaml:"logo" json:"logo"`
	}	`yaml:"info"`
	Server struct{
		Port 			string 		`yaml:"port"`//端口
		Search			bool		`yaml:"search" json:"search"`//允许搜索
		Static			string		`yaml:"static"`
	}	`yaml:"server"`
	AliDrive struct{
		ApiUrl			string		`yaml:"api_url"`//阿里云盘api
		RootFolder		string		`yaml:"root_folder"`//根目录id
		//Authorization	string		`yaml:"authorization"`//授权token
		LoginToken		string		`yaml:"login_token"`
		AccessToken		string		`yaml:"access_token"`
		RefreshToken	string		`yaml:"refresh_token"`
		MaxFilesCount	int			`yaml:"max_files_count"`
	}	`yaml:"ali_drive"`
}