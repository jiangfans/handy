package request

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var (
	ErrUnauthorized = errors.New("unauthorized error")
	ErrNotFound     = errors.New("resource not found error")
)

func InvalidCodeError(reqUrl string, statusCode int) error {
	log.Errorf("request %s return invalid code %d", reqUrl, statusCode)
	switch statusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusNotFound:
		return ErrNotFound
	}

	return errors.New("invalid code error")
}
