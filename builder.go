package errors

import (
	"fmt"
)

type ErrorBuilder struct {
	code      int
	message   string
	fields    map[string][]string
	internals []Error
}

func (err *ErrorBuilder) WithInternalError(e error) *ErrorBuilder {
	if rerr, ok := e.(*Error); ok {
		if rerr.HTTPCode() == ToHttpStatus(ErrMultipleError.ErrorCode()) {
			err.internals = append(err.internals, rerr.Internals...)
			return err
		}
		err.internals = append(err.internals, *rerr)
		return err
	}
	err.internals = append(err.internals, *ToError(e))
	return err
}

func (err *ErrorBuilder) WithInternalErrors(internals []*Error) *ErrorBuilder {
	if len(internals) > 0 {
		for _, e := range internals {
			err.internals = append(err.internals, *e)
		}
	}
	return err
}

func (err *ErrorBuilder) WithField(nm string, v interface{}) *ErrorBuilder {
	if nil == err.fields {
		err.fields = map[string][]string{}
	}
	err.fields[nm] = append(err.fields[nm], fmt.Sprint(v))
	return err
}

func (err *ErrorBuilder) Fields() map[string][]string {
	return err.fields
}

func (err *ErrorBuilder) FieldsWithDefault() map[string][]string {
	if nil == err.fields {
		err.fields = map[string][]string{}
	}
	return err.fields
}

func (err *ErrorBuilder) Build() *Error {
	var fields map[string][]string
	var internals []Error
	if len(err.fields) > 0 {
		fields = err.fields
	}

	if len(err.internals) > 0 {
		internals = err.internals
	}

	return &Error{
		Code:      err.code,
		Message:   err.message,
		Fields:    fields,
		Internals: internals,
	}
}

func Build(code int, msg string) *ErrorBuilder {
	return &ErrorBuilder{
		code:    code,
		message: msg,
	}
}

func ReBuildFromRuntimeError(e RuntimeError) *ErrorBuilder {
	var fields map[string][]string
	var internals []Error
	if err, ok := e.(*Error); ok {
		if len(err.Fields) > 0 {
			fields = map[string][]string{}
			for k, v := range err.Fields {
				fields[k] = v
			}
		}

		if len(err.Internals) > 0 {
			internals = make([]Error, len(err.Internals))
			copy(internals, err.Internals)
		}
	}
	return &ErrorBuilder{
		code:      e.ErrorCode(),
		message:   e.Error(),
		fields:    fields,
		internals: internals,
	}
}

func ReBuildFromError(e error, code int) *ErrorBuilder {
	if err, ok := e.(RuntimeError); ok {
		return ReBuildFromRuntimeError(err)
	}
	if ec, ok := GetErrorCode(e); ok {
		return &ErrorBuilder{
			code:    ec,
			message: e.Error(),
		}
	}
	return &ErrorBuilder{
		code:    code,
		message: e.Error(),
	}
}

func BuildApplicationErrorFromError(e error, code int) *Error {
	return ToError(e, code)
}
