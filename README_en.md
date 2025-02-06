# FunChain

FunChain is a lightweight Go library designed for elegant function chaining with built-in error handling, deferred resource cleanup, and hook functions.

## Features

- Elegant function chaining
- Automatic parameter passing and type conversion
- Supports functions with any type and number of parameters
- Smart error handling mechanism
- Support for deferred cleanup
- Support for error handling hooks
- Type-safe operations
- Intuitive API design

## Installation

```bash
go get github.com/jiazhoulvke/funchain@latest
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/jiazhoulvke/funchain"
)

func main() {
    var result string
    results, err := funchain.New(func() int {
        return 42
    }).Then(func(n int) string {
        return fmt.Sprintf("The number is: %d", n)
    }).Do(&result)

    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println(result) // Output: The number is: 42
}
```

### Error Handling

```go
func main() {
    results, err := funchain.New(func() (int, error) {
        return 0, errors.New("an error occurred")
    }).Then(func(n int) string {
        return fmt.Sprintf("The number is: %d", n)
    }).OnError(func(args []interface{}, err error) {
        fmt.Printf("Error occurred: %v, Args: %v\n", err, args)
    }).Do()

    if err != nil {
        fmt.Println("Error:", err)
    }
}
```

### Using Defer for Resource Cleanup

```go
func main() {
    file := createTempFile()

    results, err := funchain.New(func() error {
        return processFile(file)
    }).Defer(func() {
        file.Close()
        os.Remove(file.Name())
    }).Do()
}
```

### Complex Parameter Passing

```go
func main() {
    var (
        sum     int
        message string
    )

    results, err := funchain.New(func() (int, string) {
        return 100, "initial value"
    }).Then(func(n int, s string) (int, string) {
        return n * 2, s + " processed"
    }).Do(&sum, &message)

    fmt.Printf("sum=%d, message=%s\n", sum, message)
}
```

## API Reference

### `New(fns ...interface{}) *FunChain`

Creates a new function chain. You can pass one or more functions to be executed. Each function can be of any type, with no restrictions on the number of parameters and return values. If the provided argument is not a function, it will be skipped.

```go
chain := funchain.New(
    func() int {
        return 42
    },
    func(n int) string {
        return fmt.Sprintf("The number is: %d", n)
    },
)
```

### `Then(fns ...interface{}) *FunChain`

This method appends subsequent functions for execution. In contrast to the `New` method—which initializes the function chain—`Then` is primarily designed to achieve a Promise-like effect by automatically passing the return values of one function as the arguments to the next in the chain.

Features:

- Automatic parameter passing: The return values from the previous function are directly provided as arguments to the next function, mirroring Promise chaining.
- Type matching: Ensures that the parameter types between consecutive functions are compatible.
- Supports any number of parameters and return values, offering extensive flexibility in constructing function chains.

Example:

```go
chain := funchain.New(func() int {
    return 42
}).Then(
    func(n int) string {
        return fmt.Sprintf("The number is: %d", n)
    },
    func(s string) {
        fmt.Println(s)
    },
)
```

### `Defer(fs ...func()) *FunChain`

Adds cleanup functions to be called after the function chain execution:

- Supports multiple defer functions.
- Executed in last-in-first-out (LIFO) order.
- Commonly used for resource cleanup, logging, etc.

```go
chain.Defer(cleanup1, cleanup2)
```

### `OnError(hooks ...ErrorHookFunc) *FunChain`

Adds error handling hook functions:

- Called when an error occurs.
- Can access error information and current parameters.
- Supports multiple hook functions.
- Commonly used for error logging, monitoring, etc.

```go
chain.OnError(func(args []interface{}, err error) {
    log.Printf("Error: %v, Args: %v", err, args)
})
```

### `Before(hooks ...BeforeHookFunc) *FunChain`

Adds hook functions to be called before each function execution:

- Called before each function execution.
- Can access the parameters that will be passed to the function.
- Supports multiple hook functions.
- Commonly used for parameter validation, logging, etc.

```go
chain.Before(func(args []interface{}) {
    log.Printf("About to execute function, Args: %v", args)
})
```

### `After(hooks ...AfterHookFunc) *FunChain`

Adds hook functions to be called after each function execution:

- Called after each function execution.
- Can access the return values of the function.
- Supports multiple hook functions.
- Commonly used for result validation, performance monitoring, etc.

```go
chain.After(func(args []interface{}) {
    log.Printf("Function execution completed, Return values: %v", args)
})
```

### Hook Function Usage Example

```go
func main() {
    var result int
    _, err := funchain.New(func() int {
        return 42
    }).Before(func(input []interface{}) {
        log.Println("Before function execution", input)
    }).After(func(input []interface{}, output []interface{}) {
        log.Printf("Parameters: %v, Return values: %v", input, output)
    }).Do(&result)

    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println(result) // Output: 42
}
```

### `Do(out ...interface{}) (result []interface{}, err error)`

Executes the entire function chain, supporting two ways to retrieve return values:

1. Retrieve via the Returned Results Slice:
   - The **results** slice contains all return values from the last function in the chain.
   - The returned values are stored in order.
   - Manual type assertions are required.

```go
results, err := chain.Do()
if err == nil {
    firstResult := results[0].(string)
    secondResult := results[1].(int)
}
```

2. Retrieve via Pointer Parameters:
   - Pass pointers corresponding to the number of return values.
   - The results are automatically assigned to the provided variables.
   - No manual type assertions are needed (types must match exactly).

```go
var (
    str string
    num int
)
_, err := chain.Do(&str, &num)
if err == nil {
    fmt.Printf("str=%s, num=%d\n", str, num)
}
```

Other features:

- Automatic error handling.
- Ensured execution of all deferred functions.
- If any function in the chain returns an error, execution stops immediately.
- Supports simultaneous use of both methods to retrieve return values.

## Error Handling Mechanism

1. If a function returns an error, the chain stops execution immediately.
2. Each function may return at most one error.
3. Deferred functions are executed even if an error occurs.
4. All registered error hooks are triggered when an error occurs.

## Best Practices

1. Use `defer` functions for resource cleanup.
2. Plan function parameters and return values carefully.
3. Use `OnError` for error monitoring and logging.
4. Avoid overly complex function signatures in the chain.

## Notes

- All registered deferred and hook functions are internally wrapped with panic recovery to ensure that a panic in any of them does not interrupt the execution of the chain. However, it is recommended to handle exceptions gracefully within your hook functions.
- When the number of provided arguments is insufficient for the target function, the system automatically fills the missing parameters with their corresponding zero values.

### Thread Safety

- FunChain currently does not implement concurrency safety mechanisms. If multiple goroutines use and modify the same FunChain instance simultaneously, race conditions may occur. Please add synchronization protection or avoid modifying the instance simultaneously in concurrent environments.
