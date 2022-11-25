package gorm_tools

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// 避免数据库时区问题，存储时间统一为0时区
type Time time.Time

const ISO8601 = "2006-01-02T15:04:05Z"

func (t Time) Value() (v driver.Value, err error) {
	if time.Time(t).IsZero() {
		return nil, nil
	}

	return time.Time(t).UTC().Format("2006-01-02 15:04:05.999999"), nil
}

func (t *Time) Scan(value interface{}) (err error) {
	if value == nil {
		*t = Time(time.Time{}.UTC())
	}

	var tt time.Time

	switch value.(type) {
	case time.Time:
		tt = value.(time.Time).In(time.UTC)
	case []byte:
		valueStr := string(value.([]byte))
		if valueStr == "0000-00-00 00:00:00" {
			// zero value
			tt = time.Time{}.In(time.UTC)
		} else {
			tt, err = time.ParseInLocation("2006-01-02 15:04:05.999999", valueStr, time.UTC)
			if err != nil {
				return fmt.Errorf("parse %s failed", valueStr)
			}
		}
	default:
		return fmt.Errorf("can't convert %v to time", value)
	}

	*t = Time(tt)
	return
}

func (t Time) MarshalJSON() (value []byte, err error) {
	ts := time.Time(t).UTC().Format(ISO8601)
	return []byte(fmt.Sprintf(`"%s"`, ts)), nil
}

func (t *Time) UnmarshalJSON(bytes []byte) (err error) {
	if len(bytes) == 0 || string(bytes) == `""` {
		return nil
	}

	ret, err := time.ParseInLocation(fmt.Sprintf(`"%s"`, ISO8601), string(bytes), time.UTC)
	if err != nil {
		return err
	}

	*t = Time(ret)
	return
}

func (t Time) String() string {
	return time.Time(t).UTC().Format(ISO8601)
}

func (t *Time) Unix() int64 {
	return time.Time(*t).Unix()
}

func (t *Time) IsZero() bool {
	return time.Time(*t).IsZero()
}

func (t *Time) AddDate(years int, months int, days int) Time {
	return Time(time.Time(*t).AddDate(years, months, days))
}

// _________________________________

type UUIDBinary string

func (id UUIDBinary) Value() (v driver.Value, err error) {
	if id == "" {
		return nil, nil
	}

	u, err := uuid.Parse(string(id))
	if err != nil {
		return nil, err
	}
	uuidBinary, err := u.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return uuidBinary, nil
}

func (id *UUIDBinary) Scan(value interface{}) (err error) {
	if value == nil {
		*id = ""
		return nil
	}
	bs, ok := value.([]byte)

	if !ok {
		*id = ""
		return fmt.Errorf("%v is not bytes type", value)
	}

	u, err := uuid.FromBytes(bs)
	if err != nil {
		*id = ""
		return err
	}
	*id = UUIDBinary(u.String())
	return nil
}

func NewUUIDBinary() UUIDBinary {
	return UUIDBinary(uuid.New().String())
}

// _________________________________

type StringArray []string

func (a *StringArray) Scan(src interface{}) error {
	if bs, ok := src.([]byte); ok {
		if string(bs) == "" {
			*a = make(StringArray, 0)
		} else {
			*a = strings.SplitN(string(bs), ",", -1)
		}
		return nil
	}
	return fmt.Errorf("cannot convert %T to StringArray", src)
}

func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "", nil
	}

	return strings.Join(a, ","), nil
}

// _________________________________

type Json []byte

func (j Json) Value() (driver.Value, error) {
	if j.IsNull() {
		return nil, nil
	}
	return string(j), nil
}

func (j *Json) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	s, ok := value.([]byte)

	if !ok {
		return errors.New("invalid scan source")
	}
	*j = append((*j)[0:0], s...)
	return nil
}

func (j Json) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return j, nil
}

func (j *Json) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("null point exception")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

func (j Json) IsNull() bool {
	return len(j) == 0 || string(j) == "null"
}

func (j Json) Equals(j1 Json) bool {
	return bytes.Equal(j, j1)
}

func (j Json) String() string {
	return string(j)
}
