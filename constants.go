package errors

import "net/http"

//import (
//	"net/http"
//)

var (
	ErrTimeout        = NewError(http.StatusGatewayTimeout*1000+1, "timeout")
	ErrNotFound       = NewError(http.StatusNotFound*1000, "not found")
	ErrFieldNotExists = NewError(http.StatusNotFound*1000+201, "field isnot found")
	ErrDisabled       = NewError(http.StatusForbidden*1000+1, "disabled")
	ErrNotAcceptable  = NewError(http.StatusNotAcceptable*1000+1, "not acceptable")
	ErrNotImplemented = NewError(http.StatusNotImplemented*1000+1, "not implemented ")
	ErrPending        = NewError(570*1000+1, "pending")
	ErrRequired       = NewError(http.StatusBadRequest*1000+900, "required")

	ErrNetworkError     = NewError(560000, "network error")
	ErrInterruptError   = NewError(561000, "interrupt error")
	ErrMultipleError    = NewError(562000, "multiple error")
	ErrTableIsNotExists = NewError(591000, "table isnot exists")
	ErrResultEmpty      = NewError(592000, "results is empty")
	ErrKeyNotFound      = NewError(http.StatusNotFound*1000+501, "key isnot exists")

	ErrReadResponseFail      = NewError(560011, "network error")
	ErrUnmarshalResponseFail = NewError(560012, "network error")

	ArgumentMissing = ErrRequired
	ArgumentEmpty   = NewError(http.StatusBadRequest*1000+901, "empty")
)

func ToHttpCode(code int) int {
	if code < 1000 {
		return code
	}
	return code / 1000
}
