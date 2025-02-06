package funchain

import (
	"errors"
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

// New creates a new function chain.
// fn: the initial function to be executed.
// fn can be any type of function, with no restrictions on the number of parameters and return values.
// Parameter types between functions must be compatible.
// Functions cannot return more than one error.
func New(fn interface{}) *FunChain {
	fc := &FunChain{
		funcs:       make([]interface{}, 0, 1),
		defers:      make([]func(), 0),
		beforeHooks: make([]BeforeHookFunc, 0),
		afterHooks:  make([]AfterHookFunc, 0),
		errHooks:    make([]ErrorHookFunc, 0),
	}
	fc.funcs = append(fc.funcs, fn)
	return fc
}

// Then adds the next function to be executed.
// fn: function to be executed.
// fn can be any type of function, with no restrictions on the number of parameters and return values.
// Return values from the previous function are automatically passed to the next function, except for errors.
// Parameter types between functions must be compatible.
// Functions cannot return more than one error.
func (fc *FunChain) Then(fn interface{}) *FunChain {
	fc.funcs = append(fc.funcs, fn)
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

// ErrorHook adds error handling functions.
// hooks: list of error handling functions.
func (fc *FunChain) ErrorHook(hooks ...ErrorHookFunc) *FunChain {
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
	// Use reflect.TypeOf((*error)(nil)).Elem() for robust error type checking.
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
	out := rf.Call(in)
	result := make([]interface{}, 0, 1)
	var err error
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
