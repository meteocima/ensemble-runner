package errors

import (
	"testing"

	"errors"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	testerr := errors.New("TEST ERR")
	t.Run("is err", func(t *testing.T) {
		assert.Error(t, newRunTimeFailure(testerr))
	})

	t.Run("String", func(t *testing.T) {
		expected := `TEST ERR
	func errors.TestErrors.func2 in errors/errors_test.go:20`
		actualErr := newRunTimeFailure(testerr)
		actual := actualErr.Error() + "\n" + actualErr.Stack()
		assert.Equal(t, expected, actual)

		actualErr = func() (err error) {
			defer OnFailuresSet(&err)
			FailErr(testerr)
			return nil
		}().(runTimeFailure)
		actual = actualErr.Error() + "\n" + actualErr.Stack()
		expected = `TEST ERR
	func errors.TestErrors.func2.1 in errors/errors_test.go:26
	func errors.TestErrors.func2   in errors/errors_test.go:28`
		assert.Equal(t, expected, actual)
	})

	t.Run("FailF", func(t *testing.T) {
		var actual string
		func() {
			defer OnFailuresDo(func(err RunTimeError) {
				actual = err.Message()
			})
			FailF("Test %s", "error")
		}()
		assert.Equal(t, "Test error", actual)
	})

	t.Run("Check", func(t *testing.T) {
		var actual string
		checkFn := func(err error) {
			defer OnFailuresDo(func(err RunTimeError) {
				actual = err.Message()
			})
			Check(err)
		}

		checkFn(errors.New("Testing Err"))
		assert.Equal(t, "Testing Err", actual)

		actual = "no err"
		checkFn(nil)
		assert.Equal(t, "no err", actual)
	})

	t.Run("CheckResult", func(t *testing.T) {
		var actual string
		var actualResult = int32(-1)
		checkFn := func(err error) {
			defer OnFailuresDo(func(err RunTimeError) {
				actual = err.Message()
			})
			actualResult = CheckResult(int32(42), err)
		}

		checkFn(errors.New("Testing Err"))
		assert.Equal(t, "Testing Err", actual)
		assert.Equal(t, int32(-1), actualResult)

		actual = "no err"
		actualResult = 0
		checkFn(nil)
		assert.Equal(t, "no err", actual)
		assert.Equal(t, int32(42), actualResult)
	})

	t.Run("OnFailuresSet", func(t *testing.T) {
		fail := func() (err error) {
			defer OnFailuresSet(&err)
			FailErr(errors.New("Testing Err"))
			return nil
		}
		assert.Equal(t, "Testing Err", fail().(RunTimeError).Message())
		panics := func() (err error) {
			defer OnFailuresSet(&err)
			panic(errors.New("42"))
		}
		assert.PanicsWithError(t, "42", func() {
			panics()
		})
	})

	t.Run("OnFailuresDo", func(t *testing.T) {
		fail := func() (err error) {
			defer OnFailuresDo(func(rtErr RunTimeError) { err = rtErr })
			FailErr(errors.New("Testing Err"))
			return nil
		}
		assert.Equal(t, "Testing Err", fail().(RunTimeError).Message())
		panics := func() (err error) {
			defer OnFailuresDo(func(rtErr RunTimeError) { err = rtErr })
			panic(errors.New("42"))
		}
		assert.PanicsWithError(t, "42", func() {
			panics()
		})
	})

	t.Run("OnFailuresWrap", func(t *testing.T) {
		fail := func() (err error) {
			defer OnFailuresSet(&err)
			defer OnFailuresWrap("with arg as %d %0.1f: %w", 42, 1.1)
			FailErr(errors.New("Testing Err"))
			return nil
		}
		assert.Equal(t, "with arg as 42 1.1: Testing Err", fail().(RunTimeError).Message())
		panics := func() (err error) {
			defer OnFailuresSet(&err)
			defer OnFailuresWrap("WRAPPED: %w")
			panic(errors.New("42"))
		}
		assert.PanicsWithError(t, "42", func() {
			panics()
		})
	})

	t.Run("init module failed", func(t *testing.T) {
		assert.Panics(t, func() { errorsModuleInitFailed(true) })
	})

}
