package config

import "fmt"

type NotEmptyString string

func (ak *NotEmptyString) SetValue(s string) error {
	if s == "" {
		return fmt.Errorf("scan't be empty")
	}

	*ak = NotEmptyString(s)

	return nil
}
