package errors

import (
	"database/sql"
	nerrors "errors"
	"fmt"
	"net/http"
	"strings"
)

type DetailError interface {
	GetDetails() string
}

type HTTPError interface {
	error
	HTTPCoder
}

type HTTPCoder interface {
	HTTPCode() int
}

type ErrorCoder interface {
	ErrorCode() int
}

//  RuntimeError 一个带 Code 的 error
type RuntimeError interface {
	HTTPError
	ErrorCoder
}

var _ DetailError = &Error{}
var _ RuntimeError = &Error{}
var _ HTTPError = &Error{}
var _ Wrapper = &errwrap{}
var _ Wrapper = &Error{}

// ValidationError simple struct to store the Message & Key of a validation error
type ValidationError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Error struct {
	Code      int                 `json:"code,omitempty"`
	Message   string              `json:"message"`
	Details   string              `json:"details,omitempty"`
	Cause     error               `json:"-"`
	Fields    map[string][]string `json:"data,omitempty"`
	Internals []Error             `json:"internals,omitempty"`
}

func (err *Error) Error() string {
	if err.HTTPCode() == ErrMultipleError.HTTPCode() {
		var buffer strings.Builder
		if err.Message != "" {
			buffer.WriteString(err.Message)
			if !strings.HasSuffix(err.Message, ":") {
				buffer.WriteString(":")
			}
		} else {
			buffer.WriteString("发生多个错误:")
		}
		for _, e := range err.Internals {
			buffer.WriteString("\r\n  ")
			buffer.WriteString(e.Error())
		}
		return buffer.String()
	}

	return err.Message
}

func (err *Error) Unwrap() error {
	return err.Cause
}

func (err *Error) ErrorCode() int {
	return err.Code
}

func (err *Error) HTTPCode() int {
	return ToHttpCode(err.Code)
}

func (err *Error) GetDetails() string {
	return err.Details
}

func (err *Error) WithValidationError(key string, e string) *Error {
	if err.Fields == nil {
		err.Fields = map[string][]string{}
	}
	err.Fields[key] = append(err.Fields[key], e)
	return err
}

var errMissing = nerrors.New("err is nil")

//  RuntimeWrap 为 error 增加上下文信息
func RuntimeWrap(e error, s string, args ...interface{}) RuntimeError {
	if "" == s {
		return ToRuntimeError(e)
	}

	msg := fmt.Sprintf(s, args...) + ": " + e.Error()
	if re, ok := e.(RuntimeError); ok {
		return &ApplicationError{Cause: e, Code: re.ErrorCode(), Message: msg}
	}
	if re, ok := e.(interface {
		ErrorCode() int
	}); ok {
		return &ApplicationError{Cause: e, Code: re.ErrorCode(), Message: msg}
	}
	if re, ok := e.(HTTPError); ok {
		return &ApplicationError{Cause: e, Code: re.HTTPCode(), Message: msg}
	}

	if e == sql.ErrNoRows {
		return &ApplicationError{Cause: e, Code: http.StatusNotFound, Message: msg}
	}

	return &ApplicationError{Cause: e, Code: http.StatusInternalServerError, Message: msg}
}

func Wrap(err error, msg string) error {
	if err == nil {
		panic(errMissing)
	}
	if he, ok := err.(*Error); ok {
		newErr := *he
		newErr.Message = msg + ": " + he.Message
		return &newErr
	}
	if he, ok := err.(HTTPError); ok {
		return &Error{
			Code:    he.HTTPCode(),
			Message: msg + ": " + he.Error(),
		}
	}
	return errwrap{err: err, msg: msg}
}

func Wrapf(err error, msg string, args ...interface{}) error {
	return Wrap(err, fmt.Sprintf(msg, args...))
}

func WrapWithSuffix(err error, msg string) error {
	if err == nil {
		panic(errMissing)
	}
	if he, ok := err.(*Error); ok {
		newErr := *he
		newErr.Message = he.Message + ":" + msg
		return &newErr
	}
	if he, ok := err.(HTTPError); ok {
		return &Error{
			Code:    he.HTTPCode(),
			Message: he.Error() + ": " + msg,
		}
	}
	return errwrap{err: err, msg: msg, isSuffix: true}
}

func New(msg string) error {
	return nerrors.New(msg)
}

type errwrap struct {
	err      error
	msg      string
	isSuffix bool
}

func (e errwrap) Error() string {
	if e.isSuffix {
		return e.err.Error() + ": " + e.msg
	}
	return e.msg + ": " + e.err.Error()
}

func (e errwrap) Unwrap() error {
	return e.err
}
