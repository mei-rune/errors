package errors

import (
	"database/sql"
	"net/http"
	"strings"
	"unicode"
)

func NewError(code int, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

func NewHTTPError(code int, msg string) HTTPError {
	return NewError(code, msg)
}

func Concat(list ...Error) *Error {
	return &Error{Code: ErrMultipleError.HTTPCode(), Internals: list}
}

var ErrArray = Concat

func ErrBadArgument(paramName string, value interface{}, err ...error) HTTPError {
	if len(err) == 0 {
		return &Error{Code: http.StatusBadRequest, Message: "param '" + paramName + "' is invalid"}
	}
	return &Error{Code: http.StatusBadRequest, Message: "param '" + paramName + "' is invalid - " + err[0].Error()}
}

func BadArgument(msg string) *Error {
	return NewError(http.StatusBadRequest, msg)
}

func IsUnauthorizedError(err error) bool {
	re, ok := err.(HTTPError)
	return ok && re.HTTPCode() == http.StatusUnauthorized
}

func ToError(err error, defaultCode int) *Error {
	if he, ok := err.(*Error); ok {
		return he
	}

	result := &Error{
		Code:    defaultCode,
		Message: err.Error(),
	}
	if he, ok := err.(HTTPError); ok {
		result.Code = he.HTTPCode()
	}

	for err != nil {
		if x, ok := err.(interface{ Fill(*Error) }); ok {
			x.Fill(result)
		}

		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	return result
}

func AsHTTPError(err error) (HTTPError, bool) {
	he, ok := err.(HTTPError)
	return he, ok
}

func IsPendingError(e error) bool {
	if re, ok := e.(HTTPError); ok {
		return re.HTTPCode() == ErrPending.HTTPCode()
	}
	return e == ErrPending
}

// IsTimeoutError 是不是一个超时错误
func IsTimeoutError(e error) bool {
	if he, ok := e.(HTTPError); ok {
		return he.HTTPCode() == ErrTimeout.HTTPCode()
	}

	s := e.Error()
	if pos := strings.IndexFunc(s, unicode.IsSpace); pos > 0 {
		se := s[pos+1:]
		return se == "time out" || se == "timeout"
	}
	return s == "time out" || s == "timeout"
}

// IsNotFound 是不是一个未找到错误
func IsNotFound(e error) bool {
	if e == sql.ErrNoRows {
		return true
	}
	if he, ok := e.(HTTPError); ok {
		return he.HTTPCode() == http.StatusNotFound
	}
	return false
}

func IsEmptyError(e error) bool {
	if he, ok := e.(HTTPError); ok {
		return he.HTTPCode() == ErrResultEmpty.HTTPCode()
	}

	return e.Error() == ErrResultEmpty.Error()
}
