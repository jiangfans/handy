package gorm_tools

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func IsDuplicateError(err error) bool {
	mysqlErr, ok := err.(*mysql.MySQLError)
	if ok {
		if mysqlErr.Number == 1062 {
			return true
		}
	}
	return false
}

func IsRecordNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, gorm.ErrRecordNotFound)
}

func LockTimeoutError(err error) bool {
	mysqlErr, ok := err.(*mysql.MySQLError)
	if ok {
		if mysqlErr.Number == 1205 {
			return true
		}
	}
	return false
}
