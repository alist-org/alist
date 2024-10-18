package conf

import (
	"path/filepath"

	"github.com/alist-org/alist/v3/cmd/flags"
	"github.com/alist-org/alist/v3/pkg/utils/random"
)

type Database struct {
	Type        string `json:"type" env:"TYPE"`
	Host        string `json:"host" env:"HOST"`
	Port        int    `json:"port" env:"PORT"`
	User        string `json:"user" env:"USER"`
	Password    string `json:"password" env:"PASS"`
	Name        string `json:"name" env:"NAME"`
	DBFile      string `json:"db_file" env:"FILE"`
	TablePrefix string `json:"table_prefix" env:"TABLE_PREFIX"`
	SSLMode     string `json:"ssl_mode" env:"SSL_MODE"`
	DSN         string `json:"dsn" env:"DSN"`
	Parser      string `json:"parser" env:"PARSER"`
}

type Meilisearch struct {
	Host        string `json:"host" env:"HOST"`
	APIKey      string `json:"api_key" env:"API_KEY"`
	IndexPrefix string `json:"index_prefix" env:"INDEX_PREFIX"`
}

type Scheme struct {
	Address      string `json:"address" env:"ADDR"`
	HttpPort     int    `json:"http_port" env:"HTTP_PORT"`
	HttpsPort    int    `json:"https_port" env:"HTTPS_PORT"`
	ForceHttps   bool   `json:"force_https" env:"FORCE_HTTPS"`
	CertFile     string `json:"cert_file" env:"CERT_FILE"`
	KeyFile      string `json:"key_file" env:"KEY_FILE"`
	UnixFile     string `json:"unix_file" env:"UNIX_FILE"`
	UnixFilePerm string `json:"unix_file_perm" env:"UNIX_FILE_PERM"`
}

type LogConfig struct {
	Enable     bool   `json:"enable" env:"LOG_ENABLE"`
	Name       string `json:"name" env:"LOG_NAME"`
	MaxSize    int    `json:"max_size" env:"MAX_SIZE"`
	MaxBackups int    `json:"max_backups" env:"MAX_BACKUPS"`
	MaxAge     int    `json:"max_age" env:"MAX_AGE"`
	Compress   bool   `json:"compress" env:"COMPRESS"`
}

type TaskConfig struct {
	Workers        int  `json:"workers" env:"WORKERS"`
	MaxRetry       int  `json:"max_retry" env:"MAX_RETRY"`
	TaskPersistant bool `json:"task_persistant" env:"TASK_PERSISTANT"`
}

type TasksConfig struct {
	Download TaskConfig `json:"download" envPrefix:"DOWNLOAD_"`
	Transfer TaskConfig `json:"transfer" envPrefix:"TRANSFER_"`
	Upload   TaskConfig `json:"upload" envPrefix:"UPLOAD_"`
	Copy     TaskConfig `json:"copy" envPrefix:"COPY_"`
}

type Cors struct {
	AllowOrigins []string `json:"allow_origins" env:"ALLOW_ORIGINS"`
	AllowMethods []string `json:"allow_methods" env:"ALLOW_METHODS"`
	AllowHeaders []string `json:"allow_headers" env:"ALLOW_HEADERS"`
}

type S3 struct {
	Enable bool `json:"enable" env:"ENABLE"`
	Port   int  `json:"port" env:"PORT"`
	SSL    bool `json:"ssl" env:"SSL"`
}

type Config struct {
	Force                 bool        `json:"force" env:"FORCE"`
	SiteURL               string      `json:"site_url" env:"SITE_URL"`
	Cdn                   string      `json:"cdn" env:"CDN"`
	JwtSecret             string      `json:"jwt_secret" env:"JWT_SECRET"`
	TokenExpiresIn        int         `json:"token_expires_in" env:"TOKEN_EXPIRES_IN"`
	Database              Database    `json:"database" envPrefix:"DB_"`
	Meilisearch           Meilisearch `json:"meilisearch" envPrefix:"MEILISEARCH_"`
	Scheme                Scheme      `json:"scheme"`
	TempDir               string      `json:"temp_dir" env:"TEMP_DIR"`
	BleveDir              string      `json:"bleve_dir" env:"BLEVE_DIR"`
	DistDir               string      `json:"dist_dir"`
	Log                   LogConfig   `json:"log"`
	DelayedStart          int         `json:"delayed_start" env:"DELAYED_START"`
	MaxConnections        int         `json:"max_connections" env:"MAX_CONNECTIONS"`
	TlsInsecureSkipVerify bool        `json:"tls_insecure_skip_verify" env:"TLS_INSECURE_SKIP_VERIFY"`
	Tasks                 TasksConfig `json:"tasks" envPrefix:"TASKS_"`
	Cors                  Cors        `json:"cors" envPrefix:"CORS_"`
	S3                    S3          `json:"s3" envPrefix:"S3_"`
}

func DefaultConfig() *Config {
	tempDir := filepath.Join(flags.DataDir, "temp")
	indexDir := filepath.Join(flags.DataDir, "bleve")
	logPath := filepath.Join(flags.DataDir, "log/log.log")
	dbPath := filepath.Join(flags.DataDir, "data.db")
	return &Config{
		Scheme: Scheme{
			Address:    "0.0.0.0",
			UnixFile:   "",
			HttpPort:   5244,
			HttpsPort:  -1,
			ForceHttps: false,
			CertFile:   "",
			KeyFile:    "",
		},
		JwtSecret:      random.String(16),
		TokenExpiresIn: 48,
		TempDir:        tempDir,
		Database: Database{
			Type:        "sqlite3",
			Port:        0,
			TablePrefix: "x_",
			DBFile:      dbPath,
		},
		Meilisearch: Meilisearch{
			Host: "http://localhost:7700",
		},
		BleveDir: indexDir,
		Log: LogConfig{
			Enable:     true,
			Name:       logPath,
			MaxSize:    50,
			MaxBackups: 30,
			MaxAge:     28,
		},
		MaxConnections:        0,
		TlsInsecureSkipVerify: true,
		Tasks: TasksConfig{
			Download: TaskConfig{
				Workers:        5,
				MaxRetry:       1,
				TaskPersistant: true,
			},
			Transfer: TaskConfig{
				Workers:        5,
				MaxRetry:       2,
				TaskPersistant: true,
			},
			Upload: TaskConfig{
				Workers: 5,
			},
			Copy: TaskConfig{
				Workers:        5,
				MaxRetry:       2,
				TaskPersistant: true,
			},
		},
		Cors: Cors{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"*"},
			AllowHeaders: []string{"*"},
		},
		S3: S3{
			Enable: false,
			Port:   5246,
			SSL:    false,
		},
	}
}
