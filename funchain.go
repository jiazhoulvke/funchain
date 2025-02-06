package funchain

import (
	"errors"
	"fmt"
	"reflect"
)

// FunChain is the main type that supports chaining multiple functions.
// It provides methods to add functions to the chain along with hooks and defer (cleanup) functions.
type FunChain struct {
	funcs       []interface{}
	defers      []func()
	beforeHooks []BeforeHookFunc
	afterHooks  []AfterHookFunc
	errHooks    []ErrorHookFunc
}

// ErrorHookFunc is an error handling hook function.
// output: function return values
// err: function error
type ErrorHookFunc func(output []interface{}, err error)

// BeforeHookFunc is a hook function called before each function execution.
// input: function parameters
type BeforeHookFunc func(input []interface{})

// AfterHookFunc is a hook function called after each function execution.
// input: parameters of the previous function
// output: function return values
type AfterHookFunc func(input []interface{}, output []interface{})

// New 创建一个新的函数链。
// fns: 一个或多个待执行的函数。
// 每个函数可以是任意类型，不限制参数和返回值的数量。
// 如果传入的参数不是函数，则会被直接跳过，不会报错。
func New(fns ...interface{}) *FunChain {
	fc := &FunChain{
		funcs:       make([]interface{}, 0, len(fns)),
		defers:      make([]func(), 0),
		beforeHooks: make([]BeforeHookFunc, 0),
		afterHooks:  make([]AfterHookFunc, 0),
		errHooks:    make([]ErrorHookFunc, 0),
	}
	// 检查每个传入的参数，如果是函数则加入链中，否则跳过
	for _, fn := range fns {
		if reflect.TypeOf(fn).Kind() == reflect.Func {
			fc.funcs = append(fc.funcs, fn)
		}
	}
	return fc
}

// Then adds the next function to be executed.
// fn: function to be executed.
// fn can be any type of function, with no restrictions on the number of parameters and return values.
// Return values from the previous function are automatically passed to the next function, except for errors.
// Parameter types between functions must be compatible.
// Functions cannot return more than one error.
func (fc *FunChain) Then(fns ...interface{}) *FunChain {
	for _, fn := range fns {
		if reflect.TypeOf(fn).Kind() == reflect.Func { // 检查是否为函数类型
			fc.funcs = append(fc.funcs, fn)
		}
	}
	return fc
}

// Defer adds cleanup functions to be executed after the chain completes.
// fs: list of defer functions.
func (fc *FunChain) Defer(fs ...func()) *FunChain {
	fc.defers = append(fc.defers, fs...)
	return fc
}

// Before adds hook functions to be called before each function execution.
// hooks: list of before hook functions.
func (fc *FunChain) Before(hooks ...BeforeHookFunc) *FunChain {
	fc.beforeHooks = append(fc.beforeHooks, hooks...)
	return fc
}

// After adds hook functions to be called after each function execution.
// hooks: list of after hook functions.
func (fc *FunChain) After(hooks ...AfterHookFunc) *FunChain {
	fc.afterHooks = append(fc.afterHooks, hooks...)
	return fc
}

// OnError adds error handling functions.
// hooks: list of error handling functions.
func (fc *FunChain) OnError(hooks ...ErrorHookFunc) *FunChain {
	fc.errHooks = append(fc.errHooks, hooks...)
	return fc
}

// Do executes the function chain.
// result: function return values
// out: uses reflection to set return values to provided pointer variables.
func (fc *FunChain) Do(out ...interface{}) (result []interface{}, err error) {
	// Register all defer functions (will execute in LIFO order)
	for _, fn := range fc.defers {
		defer func(fn func()) {
			// Protect against panic in a defer function.
			defer func() {
				if r := recover(); r != nil {
					// Optionally log the panic from defer hook.
				}
			}()
			fn()
		}(fn)
	}
	var args []interface{}
	var args2 []interface{}
	for _, fn := range fc.funcs {
		// Execute all Before hooks with recovery protection.
		for _, hook := range fc.beforeHooks {
			if hook == nil {
				continue
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Optionally log or ignore panic from before hook.
					}
				}()
				hook(args)
			}()
		}
		args2, err = execFunc(fn, args)
		// Execute all After hooks with recovery protection.
		for _, hook := range fc.afterHooks {
			if hook == nil {
				continue
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Optionally log or ignore panic from after hook.
					}
				}()
				hook(args, args2)
			}()
		}
		if err != nil {
			for _, hook := range fc.errHooks {
				if hook == nil {
					continue
				}
				func() {
					defer func() {
						if r := recover(); r != nil {
							// Optionally log or ignore panic from error hook.
						}
					}()
					hook(args2, err)
				}()
			}
			return args2, err
		}
		args = args2
	}
	for i := 0; i < len(out); i++ {
		if i >= len(args) {
			break
		}
		src := reflect.ValueOf(args[i])
		dst := reflect.ValueOf(out[i])
		if dst.Kind() == reflect.Ptr {
			dst = dst.Elem()
		}
		if !dst.CanSet() {
			continue
		}
		dst.Set(src)
	}
	return args, nil
}

// execFunc executes a function with given arguments.
// f: function to be executed.
// args: arguments to pass to the function.
// returns: function return values and an error if any.
func execFunc(f interface{}, args []interface{}) ([]interface{}, error) {
	funcType := reflect.TypeOf(f)
	if funcType.Kind() != reflect.Func {
		return nil, errors.New("not a function")
	}
	// 使用 reflect.TypeOf((*error)(nil)).Elem() 进行健壮的 error 类型判断。
	var errorType = reflect.TypeOf((*error)(nil)).Elem()
	errIndex := -1
	for i := 0; i < funcType.NumOut(); i++ {
		if funcType.Out(i) == errorType {
			if errIndex != -1 {
				return nil, errors.New("more than one error")
			}
			errIndex = i
		}
	}
	funcValue := reflect.ValueOf(f)
	rf := reflect.MakeFunc(funcType, func(callArgs []reflect.Value) []reflect.Value {
		return funcValue.Call(callArgs)
	})
	in := make([]reflect.Value, 0, funcType.NumIn())
	// Pass the return values from the previous function as arguments to the next function.
	for _, arg := range args {
		in = append(in, reflect.ValueOf(arg))
	}
	// If there are fewer arguments than parameters, create zero values for the missing ones.
	for i := len(args); i < funcType.NumIn(); i++ {
		// 此处使用 reflect.Zero 获取参数对应类型的零值，确保如果传入的参数数量不足时，自动填充默认值。
		// 例如，int 类型将补上 0，string 类型则补上 ""，从而保证函数调用的参数数量与签名一致。
		in = append(in, reflect.Zero(funcType.In(i)))
	}
	var out []reflect.Value
	var err error
	// 捕获 panic 并返回错误
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic occurred: %v", r)
			}
		}()
		out = rf.Call(in)
	}()
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, 0, 1)
	for i, returnVal := range out {
		if i == errIndex {
			if e, ok := returnVal.Interface().(error); ok {
				err = e
			}
			continue
		}
		result = append(result, returnVal.Interface())
	}
	return result, err
}
