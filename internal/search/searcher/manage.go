package searcher

type New func() Searcher

var NewMap = map[string]New{}

func RegisterSearcher(config Config, searcher New) {
	NewMap[config.Name] = searcher
}
