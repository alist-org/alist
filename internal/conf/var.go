package conf

var (
	BuiltAt    string
	GoVersion  string
	GitAuthor  string
	GitCommit  string
	Version    string = "dev"
	WebVersion string
)

var (
	Conf *Config
)

var TypesMap = make(map[string][]string)
