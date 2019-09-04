package errors

import "net/http"

//import (
//	"net/http"
//)

var (
	ErrTimeout        = NewHTTPError(http.StatusGatewayTimeout*1000+1, "timeout")
	ErrNotFound       = NewHTTPError(http.StatusNotFound*1000, "not found")
	ErrDisabled       = NewHTTPError(http.StatusForbidden*1000+1, "disabled")
	ErrNotAcceptable  = NewHTTPError(http.StatusNotAcceptable*1000+1, "not acceptable")
	ErrNotImplemented = NewHTTPError(http.StatusNotImplemented*1000+1, "not implemented ")
	ErrPending        = NewHTTPError(570*1000+1, "pending")
	ErrRequired       = NewHTTPError(http.StatusBadRequest*1000+900, "required")

	ErrNetworkError     = NewHTTPError(560000, "network error")
	ErrInterruptError   = NewHTTPError(561000, "interrupt error")
	ErrMultipleError    = NewHTTPError(562000, "multiple error")
	ErrTableIsNotExists = NewHTTPError(591000, "table isnot exists")
	ErrResultEmpty      = NewHTTPError(592000, "results is empty")
	ErrKeyNotFound      = NewHTTPError(http.StatusNotFound*1000+501, "key isnot exists")

	ErrReadResponseFail      = NewHTTPError(560011, "network error")
	ErrUnmarshalResponseFail = NewHTTPError(560012, "network error")

	ArgumentMissing = ErrRequired
	ArgumentEmpty   = NewHTTPError(http.StatusBadRequest*1000+901, "empty")
)

func ToHttpCode(code int) int {
	if code < 1000 {
		return code
	}
	return code / 1000
}
