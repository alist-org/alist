package s3

var defaultPlaceholderName = ".placeholder"

func getPlaceholderName(placeholder string) string {
	if placeholder == "" {
		return defaultPlaceholderName
	}
	return placeholder
}
