package httpserver

import "strconv"

func getStringHash(s string) int64 {
	i, err := strconv.ParseInt(s, 36, 32)
	if err != nil {
		return 0
	}
	return i
}
