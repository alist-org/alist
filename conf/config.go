package conf

type Config struct {
	Server struct{
		SiteUrl 		string 		`yaml:"site_url"`//网站url
		Port 			string 		`yaml:"port"`//端口
		Search			bool		`yaml:"search"`//允许搜索
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