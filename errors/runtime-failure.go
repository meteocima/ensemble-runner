package errors

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/DataDog/gostackparse"
)

// runTimeFailure is an error that wrap an error
// and it's then paniced. Recover functions can
// discover if an error is a runTimeFailure and acting
// appropritely to recover it, or otherwise repanic it.
//
// A stack is saved at the error creation site in order
// to display it for debugging pourposes.
type runTimeFailure struct {
	err   error
	stack []byte
}

// Extends the error interface
// with Stack and Message methods
type RunTimeError interface {
	error
	Unwrap() error
	// Returns the stack frame of the runTimeFailure,
	// formatted as a string
	Stack() string

	// Returns the message of the runTimeFailure
	Message() string
}

// Wrap err in a runTimeFailure, saving
// the call stack at caller site.
func newRunTimeFailure(err error) RunTimeError {
	return runTimeFailure{
		err:   err,
		stack: debug.Stack(),
	}
}

// Unwrap returns the original error
// contained in the runTimeFailure.
// It implements the anonymous 'unwrapper'
// interface to support errors.Is and errors.As
func (rtp runTimeFailure) Unwrap() error {
	return rtp.err
}

// Return the stack frame of the runTimeFailure,
// formatted as a string
func (rtp runTimeFailure) Stack() string {
	goroutines, _ := gostackparse.Parse(bytes.NewReader(rtp.stack))

	// take only first goroutine
	stack := goroutines[0].Stack

	// skip debug.Stack frame
	stack = stack[1:]

	// skip all frame related to errors module
	for {
		if stack[0].File != thisSrcFile && stack[0].File != errorsSrcFile && stack[0].Func != "panic" {
			break
		}
		stack = stack[1:]
	}

	// skip all frame related to testing module
	lastNonTesting := len(stack)
	for i := len(stack) - 1; i >= 0; i-- {
		if strings.HasPrefix(stack[i].Func, "testing.") {
			lastNonTesting--

		}
	}

	stack = stack[0:lastNonTesting]

	var maxFuncLen = 0
	for _, frame := range stack {
		frame.File = strings.TrimPrefix(frame.File, packageRoot)
		frame.Func = strings.TrimPrefix(frame.Func, packageName)
		if len(frame.Func) > maxFuncLen {
			maxFuncLen = len(frame.Func)
		}
	}

	var stackTrace []string
	for _, frame := range stack {
		frameS := fmt.Sprintf("\tfunc %*s in %s:%d", -maxFuncLen, frame.Func, frame.File, frame.Line)
		stackTrace = append(stackTrace, frameS)
	}
	return strings.Join(stackTrace, "\n")
}

// Return the message of the runTimeFailure
func (rtp runTimeFailure) Message() string {
	return rtp.err.Error()
}

// Return a string describing the error.
// Implements error interface.
// Stack trace is included in the returned string.
func (rtp runTimeFailure) Error() string {
	/*if rtp.err == nil {
		return "Something wrong: runTimeFailure with nil err"
	}*/
	return rtp.Message() //+ "\n" + rtp.Stack()
}

func errorsModuleInitFailed(failed bool) {
	if failed {
		panic(errors.New("cannot initialize worc.errors module"))
	}
}

// Path of this source file, used to
// improve stack traces formatting
var thisSrcFile = func() string {
	_, thisSrcFile, _, ok := runtime.Caller(0)
	errorsModuleInitFailed(!ok)
	return thisSrcFile
}()

// Path of errors.go source file, used to
// improve stack traces formatting
var errorsSrcFile = func() string {
	_, thisSrcFile, _, ok := runtime.Caller(0)
	errorsModuleInitFailed(!ok)
	return filepath.Join(thisSrcFile, "../errors.go")
}()

// Root path of the worc package, used to
// improve stack traces formatting
var packageRoot = func() string {
	_, thisSrcFile, _, ok := runtime.Caller(0)
	errorsModuleInitFailed(!ok)
	return filepath.Dir(filepath.Dir(thisSrcFile)) + "/"
}()

// Complete name of the worc package, used to
// improve stack traces formatting
var packageName = func() string {
	errorsPkg := reflect.TypeOf(runTimeFailure{}).PkgPath()
	pkgParts := strings.Split(errorsPkg, "/")
	pkgParts[len(pkgParts)-1] = ""
	return strings.Join(pkgParts, "/")
}()
