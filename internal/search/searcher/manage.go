package searcher

type New func() (Searcher, error)

var NewMap = map[string]New{}

func RegisterSearcher(config Config, searcher New) {
	NewMap[config.Name] = searcher
}
