package searcher

type New func() Searcher

var NewMap = map[string]New{}

func RegisterDriver(config Config, searcher New) {
	NewMap[config.Name] = searcher
}
