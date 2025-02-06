package funchain

import (
	"errors"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestFuncChain(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString("hello")
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	var (
		s      string
		length int
	)
	result, err := New(func() (*os.File, error) {
		return os.Open(f.Name())
	}).Defer(func() {
		os.Remove(f.Name())
	}).Then(func(f *os.File) (string, error) {
		if f != nil {
			defer f.Close()
		}
		data, err := io.ReadAll(f)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}).Then(func(s string) (string, int) {
		s = s + "world"
		return s, len(s)
	}).Do(&s, &length)
	if err != nil {
		t.Fatal("Chain execution error:", err)
	}
	if s != "helloworld" {
		t.Fatal("unexpected result string: expected 'helloworld', got", s)
	}
	if length != len(s) {
		t.Fatalf("unexpected length: expected %d, got %d", len(s), length)
	}
	if !reflect.DeepEqual(result, []interface{}{"helloworld", len("helloworld")}) {
		t.Fatalf("unexpected result values: expected ['helloworld', %d], got %v", len("helloworld"), result)
	}

	// Error handling test
	_, err = New(func() error {
		return nil
	}).Then(func() error {
		return errors.New("func2 error")
	}).ErrorHook(func(args []interface{}, err error) {
		t.Logf("Error hook called with args: %v, err: %v", args, err)
	}).Do()
	if err == nil {
		t.Fatal("expected error, but got nil")
	}
	if err.Error() != "func2 error" {
		t.Fatalf("unexpected error message: expected 'func2 error', got '%s'", err.Error())
	}
}

func TestHooks(t *testing.T) {
	var (
		beforeCalled bool
		afterCalled  bool
		result       int
	)

	_, err := New(func() int {
		return 42
	}).Before(func(args []interface{}) {
		beforeCalled = true
	}).After(func(input []interface{}, output []interface{}) {
		afterCalled = true
		if len(output) != 1 || output[0].(int) != 42 {
			t.Fatal("after hook received wrong arguments")
		}
	}).Do(&result)

	if err != nil {
		t.Fatal("Chain execution error:", err)
	}
	if !beforeCalled {
		t.Fatal("before hook was not called")
	}
	if !afterCalled {
		t.Fatal("after hook was not called")
	}
	if result != 42 {
		t.Fatalf("unexpected result: expected 42, got %d", result)
	}
}

func TestEdgeCases(t *testing.T) {
	// 1. 测试链式函数之间参数数量不匹配的情况，使用零值填充
	t.Run("MismatchedArgumentsWithZeroValues", func(t *testing.T) {
		var result int
		_, err := New(func() int {
			return 42
		}).Then(func(a, b int) int {
			// a 将为 42，b 为零值 0，故 a * b 结果为 0
			return a * b
		}).Do(&result)
		if err != nil {
			t.Fatal("Unexpected error for mismatched arguments:", err)
		}
		if result != 0 {
			t.Fatalf("Unexpected result: expected %d, got %d", 0, result)
		}
	})

	// 2. 测试一个或多个钩子或延迟函数发生 panic 的情况
	t.Run("HookPanic", func(t *testing.T) {
		var result int
		_, err := New(func() int {
			return 5
		}).Before(func(args []interface{}) {
			panic("intentional panic in before hook")
		}).Then(func(n int) int {
			return n * 3
		}).Do(&result)

		if err != nil {
			t.Fatalf("Chain execution should not return an error, but returned: %v", err)
		}
		if result != 15 {
			t.Fatalf("Incorrect result: expected 15, but got %d", result)
		}
	})

	// 3. 测试多个返回值的场景
	t.Run("MultipleReturnValues", func(t *testing.T) {
		var (
			sum     int
			message string
		)
		_, err := New(func() (int, string) {
			return 100, "initial"
		}).Then(func(n int, s string) (int, string) {
			return n * 2, s + " processed"
		}).Do(&sum, &message)

		if err != nil {
			t.Fatal("Chain execution error:", err)
		}
		if sum != 200 || message != "initial processed" {
			t.Fatalf("Unexpected results: sum=%d, message=%s", sum, message)
		}
	})
}

func TestFunctionPanic(t *testing.T) {
	// 测试函数在执行过程中发生 panic 的情况
	t.Run("FunctionPanic", func(t *testing.T) {
		var result int
		panicCalled := false

		_, err := New(func() int {
			panic("intentional panic in function")
		}).Then(func(n int) int {
			return n * 2
		}).ErrorHook(func(args []interface{}, err error) {
			panicCalled = true
			t.Logf("Caught panic: %v, Args: %v", err, args)
		}).Do(&result)

		if err == nil {
			t.Fatal("Expected error due to panic in function, but got nil")
		}
		if !panicCalled {
			t.Fatal("The error hook was not called due to panic in the function")
		}
		if result != 0 {
			t.Fatalf("Expected result to be 0, but got %d", result)
		}
	})
}
