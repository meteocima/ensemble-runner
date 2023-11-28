/*
go errors management using runTimeFailure.

Worc functions panics with error of runTimeFailure type
instead of returning error codes as it's normal in go code.

This simplify the code flow, making it simpler to read and modify.
Error values are returned anyway at API boundaries.

This module provide functions to check other functions return errors (Check* functions),
to emit runTimeFailure (Fail* functions) and to recover from runTimeFailure (OnFailures* functions).

To panic with a runTimeFailure error is called to 'fail'.
*/
package errors

import (
	"fmt"
)

// When a failure occurs, set the
// err argument to the error value if
// it's a runTimeFailure, or repanic
// the error otherwise.
//
// Must be called using `defer`.
func OnFailuresSet(err *error) {
	if e := recover(); e != nil {
		if runtimeErr, ok := e.(runTimeFailure); ok {
			*err = runtimeErr
		} else {
			panic(e)
		}
	}
}

// When a failure occurs, call fn
// passing it the failure error.
// If the recovered oanic is not a
// runTimeFailure, it will be repaniced
//
// Must be called using `defer`.
func OnFailuresDo(fn func(err RunTimeError)) {
	if e := recover(); e != nil {
		if err, ok := e.(runTimeFailure); ok {
			fn(err)
		} else {
			panic(e)
		}
	}
}

// Wraps the runTimeFailure in another
// error, using fmt.Errorf, and then panic
// with the newly wrapped errors.
//
// The wrapped error is inserted as
// last argument passed to fmt.Errorf,
// so %w must be last placeholder in the
// format string.
//
// Must be called using `defer`, and after
// OnFailuresSet or OnFailuresDo
// (if any of them is called in the same func).
func OnFailuresWrap(format string, args ...any) {
	if e := recover(); e != nil {
		if runtimeErr, ok := e.(runTimeFailure); ok {
			args = append(args, runtimeErr.Unwrap())
			FailF(format, args...)
		} else {
			panic(e)
		}
	}
}

// Panic with a runTimeFailure
// that wraps err argument.
func FailErr(err error) {
	panic(newRunTimeFailure(err))
}

/*
func RtErrF(format string, args ...any) RunTimeError {
	return newRunTimeFailure(fmt.Errorf(format, args...))
}
*/
// Panic with a runTimeFailure
// built using fmt.Errorf.
func FailF(format string, args ...any) {
	FailErr(fmt.Errorf(format, args...))
}

/*
func Errorf(format string, args ...any) RunTimeError {
	return newRunTimeFailure(fmt.Errorf(format, args...))
}
*/

// If err argument is not nil, panic with a
// runTimeFailure that wraps it.
func Check(err error) {
	if err != nil {
		FailErr(err)
	}
}

// If err argument is not nil, panic with a
// runTimeFailure that wraps it. Otherwise, return
// result argument.
func CheckResult[T any](result T, err error) T {
	Check(err)
	return result
}

func CheckFn(fn func() error) {
	Check(fn())
}
