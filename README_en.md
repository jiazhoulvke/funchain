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

### Hook Function Usage Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/jiazhoulvke/funchain"
)

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

### Using `Do(out ...interface{}) (result []interface{}, err error)`

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
3. Use `ErrorHook` for error monitoring and logging.
4. Avoid overly complex function signatures in the chain.

## Notes

- All registered deferred and hook functions are internally wrapped with panic recovery to ensure that a panic in any of them does not interrupt the execution of the chain. However, it is recommended to handle exceptions gracefully within your hook functions.
- When the number of provided arguments is insufficient for the target function, the system automatically fills the missing parameters with their corresponding zero values.
