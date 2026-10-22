package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

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
		v.RegisterTagNameFunc(func(f reflect.StructField) string {
			for _, tag := range []string{"json", "form", "uri"} {
				name := strings.Split(f.Tag.Get(tag), ",")[0]
				if name != "" && name != "-" {
					return name
				}
			}
			return f.Name
		})
	}
}

func BindErrorMessage(err error) string {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) && len(verrs) > 0 {
		fe := verrs[0]
		field := fe.Field()
		switch fe.Tag() {
		case "required":
			return fmt.Sprintf("缺少参数 %s", field)
		case "mobile":
			return fmt.Sprintf("参数 %s 不是有效的手机号", field)
		case "min", "gte", "gt":
			return fmt.Sprintf("参数 %s 不能小于 %s", field, fe.Param())
		case "max", "lte", "lt":
			return fmt.Sprintf("参数 %s 不能大于 %s", field, fe.Param())
		case "oneof":
			return fmt.Sprintf("参数 %s 只能是 %s 之一", field, fe.Param())
		default:
			return fmt.Sprintf("参数 %s 不合法", field)
		}
	}
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return fmt.Sprintf("参数 %s 类型错误", typeErr.Field)
	}
	var numErr *strconv.NumError
	if errors.As(err, &numErr) {
		return fmt.Sprintf("参数格式错误：%q 不是有效的数字", numErr.Num)
	}
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return "请求体不是有效的 JSON"
	}
	return "参数错误"
}
