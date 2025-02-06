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

func TestHookPanic(t *testing.T) {
	var result int
	hookPanicCalled := false
	chain := New(func() int {
		return 5
	}).Before(func(args []interface{}) {
		// 此 Hook 故意引发 panic
		panic("intentional panic in before hook")
	}).Before(func(args []interface{}) {
		// 在上一个 hook 发生 panic 后，此 Hook 依然应该正常调用
		hookPanicCalled = true
	}).Then(func(n int) int {
		return n * 3
	})

	_, err := chain.Do(&result)
	if err != nil {
		t.Fatalf("Chain execution error: %v", err)
	}
	if result != 15 {
		t.Fatalf("Unexpected result: expected 15, got %d", result)
	}
	if !hookPanicCalled {
		t.Fatal("The second before hook was not called due to panic in the first hook")
	}
}
