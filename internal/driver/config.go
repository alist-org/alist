package driver

type Config struct {
	Name        string
	LocalSort   bool
	OnlyLocal   bool
	OnlyProxy   bool
	NoCache     bool
	NoUpload    bool
	DefaultRoot string
}

func (c Config) MustProxy() bool {
	return c.OnlyProxy || c.OnlyLocal
}
