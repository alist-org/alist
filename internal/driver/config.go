package driver

type Config struct {
	Name      string
	LocalSort bool
	OnlyLocal bool
	OnlyProxy bool
	NoCache   bool
	NoUpload  bool
}

func (c Config) MustProxy() bool {
	return c.OnlyProxy || c.OnlyLocal
}
