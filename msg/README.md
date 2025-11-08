# 消息国际化与本地化

[![Go 参考文档](https://pkg.go.dev/badge/go-slim.dev/infra/msg.svg)](https://pkg.go.dev/go-slim.dev/infra/msg)
[![Go 代码质量](https://goreportcard.com/badge/go-slim.dev/infra/msg)](https://goreportcard.com/report/go-slim.dev/infra/msg)
[![测试状态](https://github.com/go-slim/msg/workflows/Test/badge.svg)](https://github.com/go-slim/msg/actions?query=workflow%3ATest)

一个功能全面的 Go 国际化(i18n)和本地化(l10n)工具包，提供强大的消息格式化、复数和区域设置管理功能。

## 功能特性

- 🏷️ **符合 BCP 47 标准**：完整的语言标签和区域标识符支持
- 🌍 **上下文感知**：与 `context.Context` 无缝集成，支持请求级本地化
- 🧩 **可扩展**：可插拔的打印机工厂和格式化器
- 🚀 **高性能**：高效缓存和零内存分配设计
- 🛡️ **线程安全**：支持并发访问
- 📚 **丰富的区域支持**：内置常用语言和地区支持

## 安装

```bash
go get go-slim.dev/infra/msg
```

## 快速开始

```go
package main

import (
	"context"
	"fmt"
	"log"

	"go-slim.dev/infra/msg"
)

func main() {
	// 创建带区域设置的上下文
	ctx := msg.WithLocaleContext(context.Background(), msg.ChineseSimplified)

	// 从上下文中获取区域设置
	if locale, ok := msg.GetLocaleFromContext(ctx); ok {
		fmt.Printf("当前区域设置: %s\n", locale)
	}

	// 使用简单格式化器示例
	printer := msg.NewSimplePrinter()
	message := printer.Sprintf("欢迎使用我们的应用程序！")
	fmt.Println(message)
}
```

## 核心概念

### 区域设置(Locale)

区域设置使用 BCP 47 格式（如 "zh-Hans-CN"、"en-US"）表示语言和地区的组合。本包提供：

- 预定义常用区域设置
- 解析和验证
- 语言、脚本和地区提取
- 区域设置匹配和回退

### 打印机(Printer)

打印机处理实际的消息格式化，包括：

- 基本字符串格式化
- 数字和货币格式化
- 日期和时间格式化
- 复数和性别支持

### 上下文集成

使用 Go 的 context 无缝管理区域设置和打印机实例：

```go
// 在上下文中设置区域
ctx := msg.WithLocaleContext(context.Background(), msg.ChineseSimplified)

// 从上下文中获取区域
if locale, ok := msg.GetLocaleFromContext(ctx); ok {
    // 使用区域设置
}
```

## 高级用法

### 使用 xtext 包

`xtext` 包基于 `golang.org/x/text` 实现了 `PrinterFactory` 接口，提供了完整的国际化解决方案。

#### 基本用法

```go
package main

import (
	"context"
	"log"
	"os"

	"go-slim.dev/infra/msg"
	"go-slim.dev/infra/msg/xtext"
)

func main() {
	// 创建打印机工厂并配置
	factory := xtext.NewPrinterFactory(
		xtext.BaseDir("./locales"),  // 自动从该目录加载翻译文件
		xtext.Fallback(msg.English),  // 如果找不到翻译则回退到英语
		xtext.LogFunc(func(msg string) {
			log.Printf("[xtext] %s", msg)
		}),
	)

	// 创建带打印机工厂的上下文
	ctx := msg.WithPrinterFactoryContext(context.Background(), factory)

	// 为特定区域创建打印机
	printer, err := factory.CreatePrinter(msg.ChineseSimplified)
	if err != nil {
		log.Fatalf("创建打印机失败: %v", err)
	}
	message := printer.Sprintf("welcome_message")
	log.Println(message)
}
```

#### 文件结构

```
./locales/
├── en.json
├── zh-Hans.json
└── zh-Hant.json
```

示例 `zh-Hans.json`:
```json
{
  "welcome_message": "欢迎使用我们的应用程序！",
  "user_greeting": "你好，%s！"
}
```

#### 代码生成

为了更好的类型安全和 IDE 支持，可以从翻译文件生成 Go 代码：

1. 在包中创建 `generate.go` 文件：

```go
//go:generate xtext generate -pkg myapp -o messages.gen.go
```

2. 运行生成器：

```bash
go generate ./...
```

这将生成强类型的消息键和辅助函数。

### 自定义格式化器

通过实现 `Printer` 接口创建自定义格式化器：

```go
type 我的格式化器 struct{}

func (f *我的格式化器) Sprintf(format string, args ...interface{}) string {
    // 自定义格式化逻辑
    return fmt.Sprintf("格式化: "+format, args...)
}

// 创建并使用格式化器
printer := &我的格式化器{}
// 在实际应用中，通常会创建一个工厂并设置到上下文中
// factory := xtext.NewPrinterFactory()
// factory.RegisterFormatter("myformat", printer)
// ctx := msg.WithPrinterFactoryContext(ctx, factory)
```

### 区域设置匹配

处理区域设置回退和匹配：

```go
package main

import (
	"fmt"
	"go-slim.dev/infra/msg"
)

func main() {
	supported := []msg.Locale{msg.English, msg.ChineseSimplified, msg.Spanish}
	preferred := msg.NewLocale("zh-Hans-CN")

	// 查找最佳匹配
	matched := msg.Match(preferred, supported...)
	fmt.Printf("%s 的最佳匹配: %s\n", preferred, matched)
}
```

## 最佳实践

1. **始终使用上下文**传递区域设置
2. 尽可能**缓存格式化器**
3. 使用前**验证区域设置**
4. 处理**缺失翻译的回退**
5. 使用**一致的**区域标识符

## 许可证

MIT

## 贡献

欢迎贡献代码！请随时提交 Pull Request。