package utils

import "time"

func MustParseCNTime(str string) time.Time {
	lastOpTime, _ := time.ParseInLocation("2006-01-02 15:04:05 -07", str+" +08", time.Local)
	return lastOpTime
}
