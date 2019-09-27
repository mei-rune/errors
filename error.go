package errors

import (
	nerrors "errors"
	"strings"
)

type HTTPError interface {
	error

	HTTPCode() int
}

// ValidationError simple struct to store the Message & Key of a validation error
type ValidationError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Error struct {
	Code      int                        `json:"code,omitempty"`
	Message   string                     `json:"message"`
	Details   string                     `json:"details"`
	Cause     error                      `json:"-"`
	Fields    map[string]ValidationError `json:"fields,omitempty"`
	Internals []Error                    `json:"internals,omitempty"`
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

func (err *Error) ErrorCode() int {
	return err.Code
}

func (err *Error) HTTPCode() int {
	return ToHttpCode(err.Code)
}

func Wrap(err error, msg string) error {
	if he, ok := err.(*Error); ok {
		he.Message = msg + ": " + he.Message
		return he
	}
	if he, ok := err.(HTTPError); ok {
		return &Error{
			Code:    he.HTTPCode(),
			Message: msg + ": " + he.Error(),
		}
	}
	return errwrap{err: err, msg: msg}
}

func New(msg string) error {
	return nerrors.New(msg)
}

type errwrap struct {
	err error
	msg string
}

func (e errwrap) Error() string {
	return e.msg + ": " + e.err.Error()
}

func (e errwrap) Unwrap() error {
	return e.err
}
