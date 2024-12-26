package errors

import (
	"database/sql"
	"encoding/json"
	nerrors "errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
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

// RuntimeError 一个带 Code 的 error
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

// RuntimeWrap 为 error 增加上下文信息
func RuntimeWrap(e error, s string, args ...interface{}) RuntimeError {
	if "" == s {
		return ToRuntimeError(e)
	}

	msg := fmt.Sprintf(s, args...) + ": " + e.Error()
	if re, ok := e.(RuntimeError); ok {
		return &ApplicationError{Cause: e, Code: re.ErrorCode(), Message: msg}
	}
	if ec, ok := GetErrorCode(e); ok {
		return &ApplicationError{Cause: e, Code: ec, Message: msg}
	}
	if hc, ok := GetHttpCode(e); ok {
		return &ApplicationError{Cause: e, Code: hc, Message: msg}
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
		newErr.Cause = err
		return &newErr
	}
	if hc, ok := GetHttpCode(err); ok {
		return &Error{
			Code:    hc,
			Message: msg + ": " + err.Error(),
			Cause:   err,
		}
	}
	return errwrap{err: err, msg: msg, mode: modePrefix}
}


func WrapWithMessage(err error, msg string) error {
	if err == nil {
		panic(errMissing)
	}
	if he, ok := err.(*Error); ok {
		newErr := *he
		newErr.Message = msg
		newErr.Cause = err
		return &newErr
	}
	if hc, ok := GetHttpCode(err); ok {
		return &Error{
			Code:    hc,
			Message: msg,
			Cause:   err,
		}
	}
	return errwrap{err: err, msg: msg, mode: modeTitle}
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
		newErr.Cause = err
		return &newErr
	}
	if hc, ok := GetHttpCode(err); ok {
		return &Error{
			Code:    hc,
			Message: err.Error() + ": " + msg,
			Cause:   err,
		}
	}
	return errwrap{err: err, msg: msg, mode: modeSuffix}
}

func New(msg string) error {
	return nerrors.New(msg)
}

type errwrap struct {
	err      error
	msg      string
	mode int
}

const (
	modePrefix = 0
	modeSuffix = 1
	modeTitle = 2
)

func (e errwrap) Error() string {
	if e.mode == modeSuffix {
		return e.err.Error() + ": " + e.msg
	}
	if e.mode == modePrefix {
		return e.msg + ": " + e.err.Error()
	}
	return e.msg
}

func (e errwrap) Unwrap() error {
	return e.err
}

func ToResponseError(response *http.Response) error {
	if response.Body == nil {
		return NewRuntimeError(http.StatusNoContent, "no content")
	}
	contentType := response.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/plain") {
		bs, _ := ioutil.ReadAll(response.Body)
		if len(bs) == 0 {
			return NewRuntimeError(response.StatusCode, response.Status)
		}
		return NewRuntimeError(response.StatusCode, string(bs))
	}
	var values map[string]interface{}
	decoder := json.NewDecoder(response.Body)
	decoder.UseNumber()
	err := decoder.Decode(&values)
	if err != nil {
		return Wrap(err, "read error info")
	}
	defer io.Copy(io.Discard, response.Body)

	var msg string
	for _, key := range []string{"message", "error", "msg"} {
		o := values[key]
		if o == nil {
			continue
		}
		msg, _ = o.(string)
		if msg != "" {
			break
		}
	}
	if msg == "" {
		msg = fmt.Sprintf("%#v", values)
	}

	e := &Error{
		Code:    response.StatusCode,
		Message: msg,
		// Details   string              `json:"details,omitempty"`
		// Cause     error               `json:"-"`
		// Fields    map[string][]string `json:"data,omitempty"`
		// Internals []Error             `json:"internals,omitempty"`
	}
	if len(values) > 0 {
		if e.Fields == nil {
			e.Fields = map[string][]string{}
		}
		for key, value := range values {
			if list, ok := value.([]string); ok {
				ss := make([]string, len(list))
				for idx := range list {
					ss[idx] = fmt.Sprintf("%#v", list[idx])
				}
				e.Fields[key] = ss
			} else {
				e.Fields[key] = []string{fmt.Sprintf("%#v", value)}
			}
		}
	}

	if v := values["code"]; v != nil {
		switch value := v.(type) {
		case json.Number:
			i, err := strconv.Atoi(value.String())
			if err == nil {
				e.Code = i
			}
		case int32:
			e.Code = int(value)
		case int64:
			e.Code = int(value)
		case int:
			e.Code = value
		case uint32:
			e.Code = int(value)
		case uint64:
			e.Code = int(value)
		case uint:
			e.Code = int(value)
		}
	}
	return e
}

func GetErrorCode(target error) (int, bool) {
  if target == nil {
    return 0, false
  }
  wc, ok := target.(ErrorCoder)
  if ok {
    return wc.ErrorCode(), true
  }
  inner, ok := wc.(Wrapper)
  if ok {
    return GetErrorCode(inner.Unwrap())
  }
  return 0, false
}

func GetHttpCode(target error) (int, bool) {
  if target == nil {
    return 0, false
  }
  hc, ok := target.(HTTPCoder)
  if ok {
    return hc.HTTPCode(), true
  }
  wc, ok := target.(ErrorCoder)
  if ok {
    return ToHttpCode(wc.ErrorCode()), true
  }
  inner, ok := wc.(Wrapper)
  if ok {
    return GetHttpCode(inner.Unwrap())
  }
  if target == sql.ErrNoRows {
		return http.StatusNotFound, true
	}
  return 0, false
}