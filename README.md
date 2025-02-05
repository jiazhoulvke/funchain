# funchain 使用说明

## 简介

funchain 是一个 javascript promise 风格的函数链式调用库，用于简化函数的链式调用和错误处理。

## 安装

```bash
go get github.com/jiazhoulvke/funchain@latest
```

## 使用

### 1. 创建 FunChain 实例

```go
fc := funchain.New(func() (string, error) {
    return "hello", nil
})
```

### 2. 链式调用

```go
fc = fc.Then(func(s string) (string, error) {
    return s + "world", nil
})
```

### 3. 执行函数链

```go
args, err := fc.Do()
if err != nil {
    // 处理错误
}
// 处理返回值
```

### 4. 错误处理

```go
fc = funchain.New(func() error {
    return errors.New("错误")
}).Then(func() error {
    // 这个函数不会被执行
    return nil
})
args, err := fc.Do()
if err != nil {
    // 处理错误
}
```

## 例子

以下是一个完整的例子：

```go
package main

import (
    "fmt"

    "github.com/jiazhoulvke/funchain"
)

func main() {
    fc := funchain.New(func() (string, error) {
        return "hello", nil
    }).Then(func(s string) (string, error) {
        return s + "world", nil
    })
    args, err := fc.Do()
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(args) // 输出：[helloworld]
}
```

## 注意事项

- 每个函数的返回值类型必须与下一个函数的参数类型匹配。
- 如果上一个函数返回的参数个数少于下一个函数的个数，则会用零值代替。
- 如果函数链中任何一个函数返回错误，整个函数链都会停止执行，并返回错误。
