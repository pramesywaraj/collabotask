package validator

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func Struct(s interface{}) error {
	return validate.Struct(s)
}

func AtLeastOneProvided(ptrs ...interface{}) bool {
	for _, p := range ptrs {
		if p == nil {
			continue
		}

		v := reflect.ValueOf(p)
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			return true
		}
	}

	return false
}
