package conf

type Database struct {
	Type        string `json:"type" env:"DB_TYPE"`
	Host        string `json:"host" env:"DB_HOST"`
	Port        int    `json:"port" env:"DB_PORT"`
	User        string `json:"user" env:"DB_USER"`
	Password    string `json:"password" env:"DB_PASS"`
	Name        string `json:"name" env:"DB_NAME"`
	DBFile      string `json:"db_file" env:"DB_FILE"`
	TablePrefix string `json:"table_prefix" env:"DB_TABLE_PREFIX"`
	SslMode     string `json:"ssl_mode" env:"DB_SLL_MODE"`
}

type Scheme struct {
	Https    bool   `json:"https" env:"HTTPS"`
	CertFile string `json:"cert_file" env:"CERT_FILE"`
	KeyFile  string `json:"key_file" env:"KEY_FILE"`
}

type CacheConfig struct {
	Expiration      int64 `json:"expiration" env:"CACHE_EXPIRATION"`
	CleanupInterval int64 `json:"cleanup_interval" env:"CLEANUP_INTERVAL"`
}

type LogConfig struct {
	Enable        bool   `json:"enable" env:"log_enable"`
	Path          string `json:"path" env:"LOG_PATH"`
	Name          string `json:"name" env:"LOG_NAME"`
	RotationTime  uint   `json:"rotation_time" env:"LOG_TIME"`
	RotationCount uint   `json:"rotation_count" env:"LOG_COUNT"`
}

type Config struct {
	Force    bool        `json:"force"`
	Address  string      `json:"address" env:"ADDR"`
	Port     int         `json:"port" env:"PORT"`
	Assets   string      `json:"assets" env:"ASSETS"`
	Database Database    `json:"database"`
	Scheme   Scheme      `json:"scheme"`
	Cache    CacheConfig `json:"cache"`
	TempDir  string      `json:"temp_dir" env:"TEMP_DIR"`
	Log      LogConfig   `json:"log"`
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
		},
		Cache: CacheConfig{
			Expiration:      60,
			CleanupInterval: 120,
		},
		Log: LogConfig{
			Enable:        true,
			Path:          "log/%Y-%m-%d-%H:%M.log",
			Name:          "log/log.log",
			RotationTime:  24,
			RotationCount: 5,
		},
	}
}
