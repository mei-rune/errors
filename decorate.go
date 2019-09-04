package errors

import (
	nerrors "errors"
	"net/http"
)

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
	e.Details = e.Message
	e.Message = title
	return e
}
