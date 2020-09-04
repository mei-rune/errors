package errors

import (
	"net/http"
)

//import (
//	"net/http"
//)

var (
	ErrTimeout        = NewError(http.StatusGatewayTimeout*1000+1, "timeout")
	ErrNotFound       = NewError(http.StatusNotFound*1000, "not found")
	ErrFieldNotExists = NewError(http.StatusNotFound*1000+201, "field isnot found")
	ErrKeyNotFound    = NewError(http.StatusNotFound*1000+501, "key isnot exists")
	ErrRecordNotFound = NewError(http.StatusNotFound*1000+202, "record isnot found")
	ErrValueNotFound  = NewError(http.StatusNotFound*1000+203, "value isnot found")
	ErrDisabled       = NewError(http.StatusForbidden*1000+1, "disabled")
	ErrNotAcceptable  = NewError(http.StatusNotAcceptable*1000+1, "not acceptable")
	ErrNotImplemented = NewError(http.StatusNotImplemented*1000+1, "not implemented ")
	ErrPending        = NewError(570*1000+1, "pending")
	ErrRequired       = NewError(http.StatusBadRequest*1000+900, "required")
	ErrPermission     = NewError(http.StatusUnauthorized*1000+101, "permission denied")
	ErrUnauthorized   = NewError(http.StatusUnauthorized*1000+102, "user is unauthorized")

	ErrTypeError      = NewError(460*1000, "type error")
	ErrValueNull      = NewError(461*1000, "value is null")
	ErrNetworkError   = NewError(560000, "network error")
	ErrInterruptError = NewError(561000, "interrupt error")
	ErrMultipleError  = NewError(562000, "multiple error")
	ErrTableNotExists = NewError(591000, "table isnot exists")
	ErrResultEmpty    = NewError(592000, "results is empty")
	ErrMultipleValues = NewError(593000, "Multiple values meet the conditions")
	ErrIDNotExists    = Required("id")
	ErrBodyNotExists  = Required("body")
	ErrBodyEmpty      = NewError(594000, "results is empty")

	ErrReadResponseFail      = NewError(560011, "network error")
	ErrUnmarshalResponseFail = NewError(560012, "network error")

	ErrBadArgument     = NewError(http.StatusBadRequest*1000, "bad argument")
	ErrArgumentMissing = ErrRequired
	ErrArgumentEmpty   = NewError(http.StatusBadRequest*1000+901, "empty")
	ErrValidationError = NewError(http.StatusBadRequest*1000+902, "bad argument")
	ErrNoContent       = NewError(http.StatusNoContent*1000+001, "no content")

	ArgumentMissing = ErrArgumentMissing
	ArgumentEmpty   = ErrArgumentEmpty

	ErrStopped = New("stopped")
)

func ToHttpCode(code int) int {
	if code < 1000 {
		return code
	}
	return code / 1000
}

func ToHttpStatus(code int) int {
	return ToHttpCode(code)
}
