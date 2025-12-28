package api

import (
	"github.com/codercollo/simple_bank/util"
	"github.com/go-playground/validator/v10"
)

// validCurrency validates supported currency values
var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		return util.IsSupportedCurrency(currency)
	}
	return false

}
