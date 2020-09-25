// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

import (
	emperrors "emperror.dev/errors"
)

// An Wrapper provides context around another error.
type Wrapper interface {
	// Unwrap returns the next error in the error chain.
	// If there is no next error, Unwrap returns nil.
	Unwrap() error
}

// Opaque returns an error with the same error formatting as err
// but that does not match err and cannot be unwrapped.
func Opaque(err error) error {
	return noWrapper{err}
}

type noWrapper struct {
	error
}

// Unwrap returns the next error in err's chain.
// If there is no next error, Unwrap returns nil.
func Unwrap(err error) error {
	return emperrors.Unwrap(err)
}

// Is returns true if any error in err's chain matches target.
//
// An error is considered to match a target if it is equal to that target or if
// it implements an Is method such that Is(target) returns true.
func Is(err, target error) bool {
	return emperrors.Is(err, target)
}

// As finds the first error in err's chain that matches a type to which target
// points, and if so, sets the target to its value and reports success. An error
// matches a type if it is of the same type, or if it has an As method such that
// As(target) returns true. As will panic if target is nil or not a pointer.
//
// The As method should set the target to its value and report success if err
// matches the type to which target points and report success.
func As(err error, target interface{}) bool {
	return emperrors.As(err, target)
}
