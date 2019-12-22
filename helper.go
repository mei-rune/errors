package errors

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"unicode"
)

func NewError(code int, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

func NewApplicationError(code int, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

func NewInternalError(msg string) *Error {
	return &Error{Code: http.StatusInternalServerError, Message: msg}
}

func NewRuntimeError(code int, msg string) RuntimeError {
	return &Error{Code: code, Message: msg}
}

func NewHTTPError(code int, msg string) HTTPError {
	return NewError(code, msg)
}

func NewTypeError(msg string, err ...error) *Error {
	return NewError(ErrTypeError.ErrorCode(), msg)
}

func Concat(list ...Error) *Error {
	return &Error{Code: ErrMultipleError.HTTPCode(), Internals: list}
}

func ErrArray(list ...interface{}) error {
	var errList []Error

	if len(list) == 0 {
		return nil
	}

	var message string
	for vidx, value := range list {
		switch values := value.(type) {
		case []interface{}:
			if len(list) == 1 {
				if len(values) == 0 {
					return nil
				}
				if len(values) == 1 {
					return values[0].(error)
				}
			}
			if len(values) == 0 {
				break
			}

			if errList == nil {
				errList = make([]Error, 0, len(values))
			}
			for idx := range values {
				errList = append(errList, *ToError(values[idx].(error)))
			}
		case []error:
			if len(list) == 1 {
				if len(values) == 0 {
					return nil
				}
				if len(values) == 1 {
					return values[0]
				}
			}
			if len(values) == 0 {
				break
			}

			if errList == nil {
				errList = make([]Error, 0, len(values))
			}
			for idx := range values {
				errList = append(errList, *ToError(values[idx]))
			}
		case []HTTPError:
			if len(list) == 1 {
				if len(values) == 0 {
					return nil
				}
				if len(values) == 1 {
					return values[0]
				}
			}
			if len(values) == 0 {
				break
			}
			if errList == nil {
				errList = make([]Error, 0, len(values))
			}
			for idx := range values {
				errList = append(errList, *ToError(values[idx]))
			}
		case []Error:
			if len(list) == 1 {
				if len(values) == 0 {
					return nil
				}
				if len(values) == 1 {
					return &values[0]
				}
			}
			if len(values) == 0 {
				break
			}
			if errList == nil {
				errList = values
			} else {
				errList = append(errList, values...)
			}
		case []*Error:
			if len(list) == 1 {
				if len(values) == 0 {
					return nil
				}
				if len(values) == 1 {
					return values[0]
				}
			}
			if len(values) == 0 {
				break
			}
			if errList == nil {
				errList = make([]Error, 0, len(values))
			}

			for _, err := range values {
				errList = append(errList, *err)
			}
		default:
			err, ok := value.(error)
			if ok {
				if len(list) == 1 {
					return err
				}
				if errList == nil {
					errList = make([]Error, 0, len(list))
				}
				errList = append(errList, *ToError(err))
				break
			}

			msg, ok := value.(string)
			if ok {
				message = msg
				break
			}
			panic(fmt.Errorf("list %d isnot error - %T", vidx, value))
		}
	}

	if len(errList) == 0 {
		return nil
	}

	if len(errList) == 1 {
		return &errList[0]
	}
	return &Error{Code: ErrMultipleError.HTTPCode(), Message: message, Internals: errList}
}

func BadArgument(paramName string, value interface{}, err ...error) HTTPError {
	if len(err) == 0 {
		return &Error{Code: http.StatusBadRequest, Message: "param '" + paramName + "' is invalid"}
	}
	return &Error{Code: http.StatusBadRequest, Message: "param '" + paramName + "' is invalid - " + err[0].Error()}
}

func BadArgumentWithMessage(msg string) *Error {
	return NewError(http.StatusBadRequest, msg)
}

//  NotFound 创建一个 ErrNotFound
func NotFound(id interface{}, typ ...string) *Error {
	if len(typ) == 0 {
		if id == nil {
			return ErrNotFound
		}

		return NewError(ErrNotFound.Code, "record with id is '"+fmt.Sprint(id)+"' isn't found")
	}

	return NewError(ErrNotFound.Code, "record with type is '"+typ[0]+"' and id is '"+fmt.Sprint(id)+"' isn't found")
}

//  NotFound 创建一个 ErrNotFound
func ErrNotFoundWith(typeName string, id interface{}) *Error {
	return NewError(http.StatusNotFound, "record with type is '"+typeName+"' and id is '"+fmt.Sprint(id)+"' isn't found")
}

//  NotFound 创建一个 ErrNotFound
func ErrNotFoundWithText(msg string) *Error {
	if msg == "" {
		return NewError(http.StatusNotFound, "not found")
	}

	return NewError(http.StatusNotFound, msg)
}

func GetDetails(err error) string {
	if o, ok := err.(DetailError); ok {
		return o.GetDetails()
	}
	return ""
}

func IsUnauthorizedError(err error) bool {
	re, ok := err.(HTTPError)
	return ok && re.HTTPCode() == http.StatusUnauthorized
}

func ToError(err error, defaultCode ...int) *Error {
	if he, ok := err.(*Error); ok {
		return he
	}

	errCode := http.StatusInternalServerError
	if len(defaultCode) > 0 {
		errCode = defaultCode[0]
	}

	result := &Error{
		Code:    errCode,
		Message: err.Error(),
		Cause:   err,
	}
	if he, ok := err.(HTTPError); ok {
		result.Code = he.HTTPCode()
	} else if err == sql.ErrNoRows {
		result.Code = http.StatusNotFound
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

func HTTPCode(err error, statusCode ...int) int {
	code := http.StatusInternalServerError
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	he, ok := err.(HTTPError)
	if ok {
		code = he.HTTPCode()
	}
	return code
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

func RecordNotFound(id interface{}) error {
	return NewApplicationError(ErrRecordNotFound.ErrorCode(), "'"+fmt.Sprint(id)+"' is not found.")
}

func IsRecordNotFoundNotExists(err error) bool {
	if he, ok := err.(ErrorCoder); ok {
		return he.ErrorCode() == ErrRecordNotFound.ErrorCode()
	}
	return false
}

func FieldNotExists(field string) error {
	return NewError(ErrFieldNotExists.ErrorCode(), "field '"+field+"' is not exists").
		WithValidationError("field", Validation.Required(nil))
}

func IsFieldNotExists(err error) bool {
	if he, ok := err.(ErrorCoder); ok {
		return he.ErrorCode() == ErrFieldNotExists.ErrorCode()
	}
	return false
}

func Required(name string) error {
	return NewError(ErrNotFound.ErrorCode(), "'"+name+"' is required.")
}

func IsTypeError(err error) bool {
	if he, ok := err.(HTTPCoder); ok {
		return he.HTTPCode() == ErrTypeError.HTTPCode()
	}
	return false
}
