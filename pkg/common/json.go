package common

import (
	"bytes"
	"database/sql/driver"
	"errors"
)

type JSON []byte

func (j JSON) IsNull() bool {
	return len(j) == 0 || string(j) == "null"
}

func (j JSON) Equals(x JSON) bool {
	return bytes.Equal([]byte(j), []byte(x))
}

func (j JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return j, nil
}

func (j *JSON) UnmarshalJSON(src []byte) error {
	if j == nil {
		return errors.New("null point exception")
	}
	*j = append((*j)[0:0], src...)
	return nil
}

func (j *JSON) Scan(src interface{}) error {
	if src == nil {
		*j = nil
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return errors.New("invalid scan source")
	}

	*j = append((*j)[0:0], b...)
	return nil
}

func (j JSON) Value() (driver.Value, error) {
	if j.IsNull() {
		return nil, nil
	}
	return string(j), nil
}

func (j JSON) String() string {
	if !j.IsNull() {
		return string(j)
	}

	return ""
}
