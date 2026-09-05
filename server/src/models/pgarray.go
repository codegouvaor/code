package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "{}", nil
	}
	return "{" + strings.Join(a, ",") + "}", nil
}

func (a *StringArray) Scan(src interface{}) error {
	if src == nil {
		*a = nil
		return nil
	}
	switch v := src.(type) {
	case []byte:
		return a.scanString(string(v))
	case string:
		return a.scanString(v)
	case []string:
		*a = v
		return nil
	}
	return fmt.Errorf("unsupported type: %T", src)
}

func (a *StringArray) scanString(s string) error {
	s = strings.TrimSpace(s)
	if s == "" || s == "{}" {
		*a = StringArray{}
		return nil
	}
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	if s == "" {
		*a = StringArray{}
		return nil
	}
	parts := strings.Split(s, ",")
	result := make(StringArray, len(parts))
	for i, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, "\"")
		result[i] = p
	}
	*a = result
	return nil
}

func (a StringArray) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(a))
}

func (a *StringArray) UnmarshalJSON(data []byte) error {
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*a = arr
	return nil
}
