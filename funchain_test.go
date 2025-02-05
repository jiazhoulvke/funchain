package funchain

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestFunChain(t *testing.T) {
	args, err := New(func() (string, error) {
		return "hello", nil
	}).Do()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("args: %v", args)
	if !reflect.DeepEqual(args, []interface{}{"hello"}) {
		t.Fatal(errors.New("wrong result"))
	}

	args, err = New(func() (string, error) {
		return "hello", nil
	}).Then(func(s string) (string, error) {
		return s + "world", nil
	}).Do()
	t.Logf("args: %v", args)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(args, []interface{}{"helloworld"}) {
		t.Fatal(errors.New("wrong result"))
	}

	args, err = New(func() error {
		return nil
	}).Then(func(arg1 interface{}) error {
		fmt.Printf("arg1: %#v\n", arg1)
		return nil
	}).Do()
	t.Logf("args: %v", args)
	if err != nil {
		t.Fatal(err)
	}

	args, err = New(func() error {
		return nil
	}).Then(func(n int) error {
		fmt.Printf("n == %d\n", n)
		return nil
	}).Do()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(args, []interface{}{}) {
		t.Fatal(errors.New("wrong result"))
	}
}
