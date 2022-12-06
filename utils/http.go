package utils

import (
	"strconv"
	"strings"
)

func Is2xxStatusCode(statusCode int) bool {
	return strings.HasPrefix(strconv.Itoa(statusCode), "2")
}
