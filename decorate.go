package errors

import (
	nerrors "errors"
	"net/http"
)

type ApplicationError = Error

func WithHTTPCode(err error, httpCode int) HTTPError {
	if err == nil {
		panic(nerrors.New("err is nil"))
	}
	e := ToError(err, httpCode)
	e.Code = httpCode
	return e
}

func WithTitle(err error, title string) error {
	if err == nil {
		panic(nerrors.New("err is nil"))
	}
	e := ToError(err, http.StatusInternalServerError)
	e.Details = e.Message
	e.Message = title
	return e
}

func ToApplicationError(err error, defaultCode ...int) *Error {
	return ToError(err, defaultCode...)
}

// ToRuntimeError 转换成 RuntimeError
func ToRuntimeError(e error, code ...int) RuntimeError {
	if re, ok := e.(RuntimeError); ok {
		return re
	}
	return ToError(e, code...)
}
