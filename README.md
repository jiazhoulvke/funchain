# FunChain

[English](https://github.com/jiazhoulvke/funchain/blob/master/README_en.md) | 简体中文

FunChain 是一个轻量级的 Go 语言函数链式调用工具库，它让你能够以优雅的方式组织和执行一系列相关函数。通过 FunChain，你可以轻松实现函数间的数据流转、错误处理和资源清理。

## 特性

- 优雅的链式函数调用
- 自动的参数传递和类型转换
- 支持任意类型和任意数量参数的函数
- 智能的错误处理机制
- 支持 defer 延迟清理
- 支持错误处理钩子
- 完全类型安全
- 直观的 API 设计

## 安装

```bash
go get github.com/jiazhoulvke/funchain@latest
```

## 快速开始

### 基础用法

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
        return fmt.Sprintf("数字是：%d", n)
    }).Do(&result)

    fmt.Println(result) // 输出: 数字是：42
}
```

### 错误处理

```go
func main() {
    results, err := funchain.New(func() (int, error) {
        return 0, errors.New("发生错误")
    }).Then(func(n int) string {
        return fmt.Sprintf("数字是：%d", n)
    }).ErrorHook(func(args []interface{}, err error) {
        fmt.Printf("发生错误：%v，参数：%v\n", err, args)
    }).Do()

    if err != nil {
        fmt.Println("错误：", err)
    }
}
```

### 使用 Defer 进行资源清理

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

### 复杂参数传递

```go
func main() {
    var (
        sum     int
        message string
    )

    results, err := funchain.New(func() (int, string) {
        return 100, "初始值"
    }).Then(func(n int, s string) (int, string) {
        return n * 2, s + " 已处理"
    }).Do(&sum, &message)

    fmt.Printf("sum=%d, message=%s\n", sum, message)
}
```

## API 详解

### `New(fn interface{}) *FunChain`

创建新的函数链。参数 `fn` 可以是任意类型的函数，包括：

- 无参数函数
- 带返回值的函数
- 返回 error 的函数
- 多返回值函数

```go
chain := funchain.New(func() string { return "hello" })
```

### `Then(fn interface{}) *FunChain`

添加下一个要执行的函数。特点：

- 自动参数传递：前一个函数的返回值会自动传递给下一个函数
- 类型匹配：确保函数间参数类型兼容
- 支持任意数量的参数和返回值

```go
chain.Then(func(s string) int { return len(s) })
```

### `Defer(fs ...func()) *FunChain`

添加在函数链执行完成后需要调用的清理函数：

- 支持多个 defer 函数
- 按照后进先出（LIFO）的顺序执行
- 常用于资源清理、日志记录等

```go
chain.Defer(cleanup1, cleanup2)
```

### `ErrorHook(hooks ...ErrorHookFunc) *FunChain`

添加错误处理钩子函数：

- 在发生错误时被调用
- 可以访问错误信息和当前参数
- 支持多个钩子函数
- 常用于错误日志、监控等

```go
chain.ErrorHook(func(args []interface{}, err error) {
    log.Printf("错误：%v，参数：%v", err, args)
})
```

### `Before(hooks ...BeforeHookFunc) *FunChain`

添加在每个函数执行前调用的钩子函数：

- 在每个函数执行前被调用
- 可以访问即将传入函数的参数
- 支持多个钩子函数
- 常用于参数验证、日志记录等

```go
chain.Before(func(args []interface{}) {
    log.Printf("即将执行函数，参数：%v", args)
})
```

### `After(hooks ...AfterHookFunc) *FunChain`

添加在每个函数执行后调用的钩子函数：

- 在每个函数执行后被调用
- 可以访问函数的返回值
- 支持多个钩子函数
- 常用于结果验证、性能监控等

```go
chain.After(func(args []interface{}) {
    log.Printf("函数执行完成，返回值：%v", args)
})
```

### 钩子函数使用示例

```go
func main() {
    var result int
    _, err := funchain.New(func() int {
        return 42
    }).Before(func(input []interface{}) {
        log.Println("函数执行前", input)
    }).After(func(input []interface{}, output []interface{}) {
        log.Printf("参数: %v, 返回值：%v", input, output)
    }).Do(&result)

    fmt.Println(result) // 输出: 42
}
```

### `Do(out ...interface{}) (result []interface{}, err error)`

执行整个函数链，支持两种方式获取返回值：

1. 通过返回的 results 切片获取：
   - results 包含最后一个函数的所有返回值
   - 返回值按顺序存储在 results 切片中
   - 需要自行进行类型断言

```go
results, err := chain.Do()
if err == nil {
    firstResult := results[0].(string)
    secondResult := results[1].(int)
}
```

2. 通过 out 参数指针获取：
   - 传入与返回值数量相同的指针参数
   - 自动将结果写入指针指向的变量
   - 无需手动进行类型断言
   - 类型必须与返回值完全匹配

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

其他特性：

- 自动进行错误处理
- 确保 defer 函数的执行
- 如果函数链中任何一步返回错误，立即终止执行
- 支持同时使用两种方式获取返回值

## 错误处理机制

FunChain 的错误处理遵循以下规则：

1. 如果函数返回 error，将立即中断执行
2. 每个函数最多只能返回一个 error
3. defer 函数即使在发生错误时也会执行
4. 错误发生时会触发所有注册的错误处理钩子

## 最佳实践

1. 使用 defer 进行资源清理
2. 合理规划函数的参数和返回值
3. 使用 ErrorHook 进行错误监控和日志记录
4. 避免在链中使用过于复杂的函数签名

## 注意事项

- 所有注册的 defer 函数和 hook 函数都会在内部捕获 panic，以确保即使其中某个出现异常，也不会中断整个函数链的执行。不过建议在 hook 函数内部自行处理异常，以避免潜在问题。
- 当传入的参数数量不足以满足目标函数的要求时，系统会自动使用对应类型的零值填充。例如，假设定义了如下函数：

  ```go
  // multiply 接收两个整数参数，返回它们的乘积
  func multiply(a int, b int) int {
      return a * b
  }
  ```

  如果链中前一个函数仅返回一个整数（例如返回 7），那么在调用 multiply 时，将自动为缺失的第二个参数填充 0，即实际调用 multiply(7, 0)，结果为 0。因此，为避免出现这种非预期行为，请确保各函数之间传递的参数数量和类型完全匹配。

- 当某个函数返回错误时，后续的函数不会被执行，但所有注册的 defer 函数仍会正常执行。
- **线程安全性：** FunChain 目前未实现并发安全机制。如果多个 goroutine 同时使用和修改同一个 FunChain 实例，可能会出现竞态问题。请在并发环境下使用时自行添加同步保护或避免同时修改实例。
