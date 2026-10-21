package utils

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var mobileRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

func InitCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("mobile", func(fl validator.FieldLevel) bool {
			mobile := fl.Field().String()
			return mobileRegex.MatchString(mobile)
		})
	}
}
