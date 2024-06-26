package errors

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"unicode"
	"runtime/debug"

	"emperror.dev/emperror"
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

func NewValidationError(message string) *Error {
	return &Error{Code: ErrValidationError.ErrorCode(), Message: message}
}

func Join(err1, err2 error) error {
	if err1 == nil {
		return err2
	}
	if err2 == nil {
		return err1
	}
	return ErrArray(err1, err2)
}

func Join3(err1, err2, err3 error) error {
	if err1 == nil {
		return Join(err2, err3)
	}
	if err2 == nil {
		return Join(err1, err3)
	}
	if err3 == nil {
		return ErrArray(err1, err2)
	}
	return ErrArray(err1, err2, err3)
}

func Concat(list ...Error) *Error {
	return &Error{Code: ErrMultipleError.ErrorCode(), Internals: list}
}

func ErrorIfNotEmpty(errList []error) error {
	if len(errList) == 0 {
		return nil
	}
	if len(errList) == 1 {
		return errList[0]
	}
	return ErrArray(errList)
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
	return &Error{Code: ErrMultipleError.ErrorCode(), Message: message, Internals: errList}
}

func NewArgumentMissing(paramName string, err ...error) HTTPError {
	if len(err) == 0 {
		return &Error{Code: ErrBadArgument.ErrorCode(), Message: "param '" + paramName + "' is missing"}
	}
	return &Error{Code: ErrBadArgument.ErrorCode(), Message: "param '" + paramName + "' is missing"}
}

func BadArgument(paramName string, value interface{}, err ...error) HTTPError {
	if len(err) == 0 {
		return &Error{Code: ErrBadArgument.ErrorCode(), Message: "param '" + paramName + "' is invalid"}
	}
	return &Error{Code: ErrBadArgument.ErrorCode(), Message: "param '" + paramName + "' is invalid - " + err[0].Error()}
}

func BadArgumentWithMessage(msg string, err ...error) *Error {
	e := NewError(ErrBadArgument.ErrorCode(), msg)
	if len(err) > 0 {
		e.Cause = err[0]
	}
	return e
}

// NotFound 创建一个 ErrNotFound
func NotFound(id interface{}, typ ...string) *Error {
	if len(typ) == 0 {
		if id == nil {
			return ErrNotFound
		}

		return NewError(ErrNotFound.Code, "record with id is '"+fmt.Sprint(id)+"' isn't found")
	}

	return NewError(ErrNotFound.Code, "record with type is '"+typ[0]+"' and id is '"+fmt.Sprint(id)+"' isn't found")
}

// NotFound 创建一个 ErrNotFound
func ErrNotFoundWith(typeName string, id interface{}) *Error {
	return NewError(ErrNotFound.Code, "record with type is '"+typeName+"' and id is '"+fmt.Sprint(id)+"' isn't found")
}

// NotFound 创建一个 ErrNotFound
func ErrNotFoundWithText(msg string) *Error {
	if msg == "" {
		return NewError(ErrNotFound.Code, "not found")
	}

	return NewError(ErrNotFound.Code, msg)
}

func RecordNotFound(id interface{}) error {
	return NewError(ErrRecordNotFound.ErrorCode(), "'"+fmt.Sprint(id)+"' is not found.")
}

func GetDetails(err error) string {
	if o, ok := err.(DetailError); ok {
		return o.GetDetails()
	}
	return ""
}

func IsUnauthorizedError(err error) bool {
	hc, ok := GetHttpCode(err)
	return ok && hc == http.StatusUnauthorized
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
	if ec, ok := GetErrorCode(err); ok {
		result.Code = ec
	} else if hc, ok := GetHttpCode(err); ok {
		result.Code = hc
	} else if err == sql.ErrNoRows {
		result.Code = ErrNotFound.Code
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

func ToErrorIfNotNull(err error, defaultCode ...int) *Error {
	if err == nil {
		return nil
	}
	return ToError(err, defaultCode...)
}

func AsHTTPError(err error) (HTTPError, bool) {
	he, ok := err.(HTTPError)
	if ok {
		return he, true
	}

	hc, ok := GetHttpCode(err)
	if ok {
		return &Error{Code: hc, Message: err.Error()}, true
	}
	return he, ok
}

func HTTPCode(err error, statusCode ...int) int {
	code := http.StatusInternalServerError
	if len(statusCode) > 0 {
		code = statusCode[0]
	}

	if hc, ok := GetHttpCode(err); ok {
		code = hc
	} else if err == sql.ErrNoRows {
		code = http.StatusNotFound
	}
	return code
}

func IsPendingError(e error) bool {
	if hc, ok := GetHttpCode(e); ok {
		return hc == ErrPending.HTTPCode()
	}
	return Is(e, ErrPending)
}
func IsMultipleChoices(e error) bool {
	if hc, ok := GetHttpCode(e); ok {
		return hc == ErrMultipleChoices.HTTPCode()
	}
	return Is(e, ErrMultipleChoices)
}

// IsTimeoutError 是不是一个超时错误
func IsTimeoutError(e error) bool {
	if hc, ok := GetHttpCode(e); ok {
		return hc == ErrTimeout.HTTPCode()
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
	if Is(e, sql.ErrNoRows) {
		return true
	}
	if hc, ok := GetHttpCode(e); ok {
		return hc == http.StatusNotFound
	}
	return false
}

func IsEmptyError(e error) bool {
	if hc, ok := GetHttpCode(e); ok {
		return hc == ErrResultEmpty.HTTPCode()
	}

	return Is(e, ErrResultEmpty)
}

func IsRecordNotFoundNotExists(err error) bool {
	if ec, ok := GetErrorCode(err); ok {
		return ec == ErrRecordNotFound.ErrorCode()
	}
	return false
}

func FieldNotExists(field string) error {
	return NewError(ErrFieldNotExists.ErrorCode(), "field '"+field+"' is not exists").
		WithValidationError("field", "Reqired")
}

func IsFieldNotExists(err error) bool {
	if ec, ok := GetErrorCode(err); ok {
		return ec == ErrFieldNotExists.ErrorCode()
	}
	return false
}

func Required(name string) error {
	return NewError(ErrNotFound.ErrorCode(), "'"+name+"' is required.")
}

func IsTypeError(err error) bool {
	if hc, ok := GetHttpCode(err); ok {
		return hc == ErrTypeError.HTTPCode()
	}
	return false
}

func IsValidationError(err error) bool {
	if hc, ok := GetHttpCode(err); ok {
		return hc == ErrValidationError.HTTPCode()
	}
	return false
}

func IsNoContent(err error) bool {
	if hc, ok := GetHttpCode(err); ok {
		return hc == http.StatusNoContent
	}
	return false
}

func IsConnectError(err error) bool {
	return strings.Contains(err.Error(), "connectex:") ||
		strings.Contains(err.Error(), "connect:")
}

func IsStopped(e error) bool {
	return e == ErrStopped
}

type ErrorHandler = emperror.ErrorHandler
type ErrorHandlerContext = emperror.ErrorHandlerContext

// Panic panics if the passed error is not nil.
// If the error does not contain any stack trace, the function attaches one, starting from the frame of the
// "Panic" function call.
//
// This function is useful with HandleRecover when panic is used as a flow control tool to stop the application.
func Panic(err error) {
	emperror.Panic(err)
}

// Recover accepts a recovered panic (if any) and converts it to an error (if necessary).
func Recover(r interface{}) error {
	return emperror.Recover(r)
}

// HandleRecover recovers from a panic and handles the error.
//
//	defer emperror.HandleRecover(errorHandler)
func HandleRecover(handler ErrorHandler) {
	emperror.HandleRecover(handler)
}



// Recover accepts a recovered panic (if any) and converts it to an error (if necessary).
func HandlePanic(err *error, context string) {
	if r := recover(); r != nil {
		switch x := r.(type) {
		case string:
			*err = New(context + ": " + x)
		case error:
			*err = Wrap(x, context)
		default:
			*err = fmt.Errorf("%s: %v", context, r)
		}

		stack := debug.Stack()
		*err = &withStack{
			err: *err,
			stack: stack,
		}
	}
}

type withStack struct {
	err error
	stack []byte
}

func (w *withStack) Error() string  {
  return fmt.Sprintf("%s\r\n%s", w.err, w.stack)
}

func (w *withStack) Cause() error  { return w.err }
func (w *withStack) Unwrap() error { return w.err }
