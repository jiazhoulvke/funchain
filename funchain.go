package funchain

import (
	"errors"
	"reflect"
)

type FunChain struct {
	// err    error
	funcs  []interface{}
	defers []func()
}

func New(f interface{}) *FunChain {
	fc := &FunChain{
		funcs:  make([]interface{}, 0, 1),
		defers: make([]func(), 0),
	}
	fc.funcs = append(fc.funcs, f)
	return fc
}

func (fc *FunChain) Then(f interface{}) *FunChain {
	fc.funcs = append(fc.funcs, f)
	return fc
}

func (fc *FunChain) Do() ([]interface{}, error) {
	var args []interface{}
	var err error
	for _, f := range fc.funcs {
		args, err = execFunc(f, args)
		if err != nil {
			return args, err
		}
	}
	return args, nil
}

// execFunc
// f: 待执行函数
// args: 上一个函数的返回值
func execFunc(f interface{}, args []interface{}) ([]interface{}, error) {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		return nil, errors.New("not a function")
	}
	errIndex := -1
	for i := 0; i < t.NumOut(); i++ {
		if t.Out(i).String() == "error" {
			if errIndex != -1 {
				return nil, errors.New("more than one error")
			}
			errIndex = i
		}
	}
	v := reflect.ValueOf(f)
	rf := reflect.MakeFunc(t, func(args []reflect.Value) (results []reflect.Value) {
		return v.Call(args)
	})
	in := make([]reflect.Value, 0, t.NumIn())
	// 将上一个函数的返回值作为下一个函数的参数
	for _, arg := range args {
		in = append(in, reflect.ValueOf(arg))
	}
	// 假如上一个函数的返回值个数少于下一个函数的参数个数,补全缺失的参数
	for i := len(args); i < t.NumIn(); i++ {
		in = append(in, reflect.New(t.In(i)).Elem())
	}
	out := rf.Call(in)
	result := make([]interface{}, 0, 1)
	var err error
	for i := 0; i < len(out); i++ {
		if i == errIndex {
			e, ok := out[i].Interface().(error)
			if ok {
				err = e
			}
			continue
		}
		result = append(result, out[i].Interface())
	}
	return result, err
}
