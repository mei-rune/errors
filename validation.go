package errors

import "regexp"

var Validation = validation{}

type validation struct{}

// Required tests that the argument is non-nil and non-empty (if string or list)
func (v validation) Required(obj interface{}) ValidationError {
	return ValidationError{Code: "REQUIRED"}
}

func (v validation) Min(n int, min int) ValidationError {
	return ValidationError{Code: "MIN"}
}

func (v validation) Max(n int, max int) ValidationError {
	return ValidationError{Code: "MAX"}
}

func (v validation) Range(n, min, max int) ValidationError {
	return ValidationError{Code: "RANGE"}
}

func (v validation) MinSize(obj interface{}, min int) ValidationError {
	return ValidationError{Code: "MINSIZE"}
}

func (v validation) MaxSize(obj interface{}, max int) ValidationError {
	return ValidationError{Code: "MAXSIZE"}
}

func (v validation) Length(obj interface{}, n int) ValidationError {
	return ValidationError{Code: "LENGTH"}
}

func (v validation) Match(str string, regex *regexp.Regexp) ValidationError {
	return ValidationError{Code: "Match"}
}

func (v validation) Email(str string) ValidationError {
	return ValidationError{Code: "EMAIL"}
}

func (v validation) IPAddr(str string, cktype ...int) ValidationError {
	return ValidationError{Code: "IP"}
}

func (v validation) MacAddr(str string) ValidationError {
	return ValidationError{Code: "MAC"}
}

func (v validation) URL(str string) ValidationError {
	return ValidationError{Code: "URL"}
}

func (v validation) Integer(str string) ValidationError {
	return ValidationError{Code: "INT"}
}

func (v validation) Datetime(str string) ValidationError {
	return ValidationError{Code: "TIME"}
}
