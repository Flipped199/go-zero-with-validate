package httpx

import (
	"context"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
	"reflect"
)

const xForwardedFor = "X-Forwarded-For"

// GetFormValues returns the form values.
func GetFormValues(r *http.Request) (map[string]any, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		if err != http.ErrNotMultipart {
			return nil, err
		}
	}

	params := make(map[string]any, len(r.Form))
	for name := range r.Form {
		formValue := r.Form.Get(name)
		if len(formValue) > 0 {
			params[name] = formValue
		}
	}

	return params, nil
}

// GetRemoteAddr returns the peer address, supports X-Forward-For.
func GetRemoteAddr(r *http.Request) string {
	v := r.Header.Get(xForwardedFor)
	if len(v) > 0 {
		return v
	}

	return r.RemoteAddr
}

type Validator struct {
	Validator *validator.Validate
	Uni       *ut.UniversalTranslator
	Trans     ut.Translator
}

func NewValidator() *Validator {
	v := Validator{}
	zh := zh.New()
	v.Uni = ut.New(zh)
	v.Validator = validator.New()
	v.Trans, _ = v.Uni.GetTranslator("zh")
	err := zh_translations.RegisterDefaultTranslations(v.Validator, v.Trans)
	if err != nil {
		logx.Errorf("校验翻译器注册失败: %s", err.Error())
		return nil
	}
	return &v
}

func (v *Validator) Validate(context context.Context, data any) string {
	if err := v.Validator.StructCtx(context, data); err != nil {
		r := reflect.TypeOf(data).Elem()
		if errs, ok := err.(validator.ValidationErrors); ok {

			if field, ok := r.FieldByName(errs[0].StructField()); ok {
				msg := field.Tag.Get("msg")
				label := field.Tag.Get("label")

				if label == "" {
					label = errs[0].StructField()
				}
				if msg == "" {
					msg = errs[0].Translate(v.Trans)[len(errs[0].StructField()):]
				}
				return label + msg
			}
		}
		invalid, ok := err.(*validator.InvalidValidationError)
		if ok {
			return invalid.Error()
		}
	}
	return ""
}
