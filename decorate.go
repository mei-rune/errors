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
	if e.Details == "" {
		e.Details = e.Message
	} else {
		e.Details = e.Message + ": " + e.Details
	}
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

type WithSQL struct {
	Err    error
	SqlStr string
	Args   []interface{}
}

func (w *WithSQL) Error() string { return w.Err.Error() }

func (w *WithSQL) Cause() error { return w.Err }

// Unwrap provides compatibility for Go 1.13 error chains.
func (w *WithSQL) Unwrap() error { return w.Err }

func WrapSQLError(err error, sqlStr string, args []interface{}) error {
	if sqlStr == "" {
		return err
	}

	return &WithSQL{Err: err, SqlStr: sqlStr, Args: args}
}

func ToSQLError(err error) *WithSQL {
	if err == nil {
		return nil
	}
	e, _ := err.(*WithSQL)
	return e
}
