package conf

type Database struct {
	Type        string `json:"type" env:"A_LIST_DB_TYPE"`
	User        string `json:"user" env:"A_LIST_DB_USER"`
	Password    string `json:"password" env:"A_LIST_DB_PASS"`
	Host        string `json:"host" env:"A_LIST_DB_HOST"`
	Port        int    `json:"port" env:"A_LIST_DB_PORT"`
	Name        string `json:"name" env:"A_LIST_DB_NAME"`
	TablePrefix string `json:"table_prefix" env:"A_LIST_DB_TABLE_PREFIX"`
	DBFile      string `json:"db_file" env:"A_LIST_DB_FILE"`
	SslMode     string `json:"ssl_mode" env:"A_LIST_SLL_MODE"`
}

type Scheme struct {
	Https    bool   `json:"https" env:"A_LIST_HTTPS"`
	CertFile string `json:"cert_file" env:"A_LIST_CERT"`
	KeyFile  string `json:"key_file" env:"A_LIST_KEY"`
}

type CacheConfig struct {
	Expiration      int64 `json:"expiration" env:"A_LIST_DB_EXPIRATION"`
	CleanupInterval int64 `json:"cleanup_interval" env:"A_LIST_CLEANUP_INTERVAL"`
}

type Config struct {
	Force    bool        `json:"force"`
	Address  string      `json:"address" env:"A_LIST_ADDR"`
	Port     int         `json:"port" env:"A_LIST_PORT"`
	Assets   string      `json:"assets" env:"A_LIST_ASSETS"`
	Database Database    `json:"database"`
	Scheme   Scheme      `json:"scheme"`
	Cache    CacheConfig `json:"cache"`
	TempDir  string      `json:"temp_dir" env:"A_LIST_TEMP_DIR"`
}

func DefaultConfig() *Config {
	return &Config{
		Address: "0.0.0.0",
		Port:    5244,
		Assets:  "https://npm.elemecdn.com/alist-web@$version/dist",
		TempDir: "data/temp",
		Database: Database{
			Type:        "sqlite3",
			Port:        0,
			TablePrefix: "x_",
			DBFile:      "data/data.db",
			SslMode:     "disable",
		},
		Cache: CacheConfig{
			Expiration:      60,
			CleanupInterval: 120,
		},
	}
}
