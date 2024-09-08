package app

import (
	"github.com/RobinHood3082/simplebank/util"
	"github.com/go-playground/validator/v10"
)

var validCurrency validator.Func = func(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		return util.IsCurrencySupported(currency)
	}
	return false
}

func SetupValidation(validate *validator.Validate) error {
	err := validate.RegisterValidation("currency", validCurrency)

	return err
}
