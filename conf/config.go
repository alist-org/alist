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

type Config struct {
	Force       bool        `json:"force"`
	Address     string      `json:"address" env:"ADDR"`
	Port        int         `json:"port" env:"PORT"`
	Assets      string      `json:"assets" env:"ASSETS"`
	LocalAssets string      `json:"local_assets" env:"LOCAL_ASSETS"`
	SubFolder   string      `json:"sub_folder" env:"SUB_FOLDER"`
	Database    Database    `json:"database"`
	Scheme      Scheme      `json:"scheme"`
	Cache       CacheConfig `json:"cache"`
	Auth        AuthConfig  `json:"auth"`
	TempDir     string      `json:"temp_dir" env:"TEMP_DIR"`
}

type AuthConfig struct {
	OrganizationName string `json:"organization_name" env:"ORGANIZATION_NAME"`
	ApplicationName  string `json:"application_name" env:"APPLICATION_NAME"`
	Endpoint         string `json:"endpoint" env:"ENDPOINT"`
	ClientId         string `json:"client_id" env:"CLIENT_ID"`
	ClientSecret     string `json:"client_secret" env:"CLIENT_SECRET"`
	JwtPublicKeyPemFile  string `json:"jwt_public_key_pem_file" env:"JWT_PUBLIC_KEY_PEM_FILE"`
}

func DefaultConfig() *Config {
	return &Config{
		Address:     "0.0.0.0",
		Port:        5244,
		Assets:      "/",
		SubFolder:   "",
		LocalAssets: "",
		TempDir:     "data/temp",
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
		Auth: AuthConfig{
			JwtPublicKeyPemFile: "token_jwt_key.pem",
		},
	}
}
