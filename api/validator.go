package api

import (
	"github.com/XiaozhouCui/go-bank/db/util"
	"github.com/go-playground/validator/v10"
)

var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	// if "ok" is true, then currency is a valid string
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		// check if currency is supported
		return util.IsSupportedCurrency(currency)
	}
	// if "ok" is false, then currency is not a string, return false
	return false
}
