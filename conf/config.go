package conf

type Database struct {
	Type        string `json:"type"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Name        string `json:"name"`
	TablePrefix string `json:"table_prefix"`
	DBFile      string `json:"db_file"`
}
type Config struct {
	Address  string   `json:"address"`
	Port     int      `json:"port"`
	Database Database `json:"database"`
}

func DefaultConfig() *Config {
	return &Config{
		Address: "0.0.0.0",
		Port:    5244,
		Database: Database{
			Type:        "sqlite3",
			User:        "",
			Password:    "",
			Host:        "",
			Port:        0,
			Name:        "",
			TablePrefix: "x_",
			DBFile:      "data/data.db",
		},
	}
}
