// Package msg 提供了国际化和本地化的核心接口和实现。
//
// 本包定义了翻译系统的抽象接口，允许不同的底层实现
// （如 golang.org/x/text、第三方库等）无缝集成。通过接口抽象，
// 应用程序可以在不修改业务逻辑的情况下切换不同的翻译引擎。
//
// 核心接口：
// - Localizer: 提供语言环境访问
// - Formatter: 提供字符串格式化功能
// - WriterFormatter: 提供写入式格式化
// - ConsoleFormatter: 提供控制台输出格式化
// - Printer: 组合所有功能的完整接口
// - PrinterFactory: 工厂模式创建 Printer 实例
//
// 设计特点：
// 1. 接口分离：不同功能分别定义接口，支持灵活组合
// 2. 工厂模式：通过工厂接口支持不同的翻译引擎
// 3. 缓存机制：内置 Printer 缓存提高性能
// 4. 线程安全：所有实现都是并发安全的
// 5. 回退支持：支持语言回退机制
//
// 使用示例：
//
//	// 基本使用
//	factory := NewPrinterFactory()
//	printer, err := factory.CreatePrinter(Chinese)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 格式化输出
//	message := printer.Sprintf("Hello, %s!", "World")
//	fmt.Println(message)
//
//	// 控制台输出
//	printer.Printf("Welcome to our application\n")
package msg

import (
	"fmt"
	"io"
)

// Localizer 定义了本地化相关的基本接口。
//
// 此接口提供对当前语言环境的访问能力。所有翻译相关的组件
// 都应该实现这个接口，以便系统能够获取和处理语言信息。
//
// 接口设计原则：
// - 最小化接口：只包含最基本的方法
// - 易于实现：任何结构体都可以轻松实现
// - 类型安全：使用强类型的 Locale 而非字符串
//
// 使用场景：
// - 获取当前使用的语言环境
// - 日志记录和调试
// - 语言偏好检测
// - 动态语言切换
type Localizer interface {
	// Locale 返回当前的语言环境。
	//
	// 返回的 Locale 应该是一个有效的 BCF 47 语言标签，
	// 如 "en", "zh-CN", "fr-FR" 等。
	//
	// 返回：
	//   Locale: 当前的语言环境
	Locale() Locale
}

// Formatter 定义了字符串格式化接口。
//
// 此接口提供了与 Go 标准库 fmt 包类似的 API，但支持国际化。
// 所有方法都返回格式化后的字符串，可以用于构建消息、日志等。
//
// 与 fmt 包的对应关系：
// - Sprint() 对应 fmt.Sprint()
// - Sprintf() 对应 fmt.Sprintf()
// - Sprintln() 对应 fmt.Sprintln()
//
// 国际化特性：
// - 根据当前语言环境调整数字、日期格式
// - 支持从翻译资源中查找本地化文本
// - 保持与标准库相同的格式化动词
type Formatter interface {
	// Sprint 使用当前语言环境格式化参数为字符串。
	//
	// 类似于 fmt.Sprint，但支持本地化格式化。
	// 所有参数按照它们的默认字符串表示进行连接。
	//
	// 参数：
	//   args: 要格式化的参数列表
	//
	// 返回：
	//   string: 格式化后的字符串
	//
	// 使用示例：
	//
	//	result := formatter.Sprint("Hello", 42, true)
	//	// 可能的输出： "Hello42true" （根据语言环境可能有不同格式）
	Sprint(args ...any) string

	// Sprintf 使用当前语言环境格式化字符串。
	//
	// 类似于 fmt.Sprintf，但支持本地化翻译和格式化。
	// 这是最常用的翻译方法，支持翻译键查找和格式化。
	//
	// 参数：
	//   format: 格式字符串，可以包含翻译键
	//   args: 要插入格式字符串的参数
	//
	// 返回：
	//   string: 本地化和格式化后的字符串
	//
	// 使用示例：
	//
	//	// 直接翻译
	//	message := formatter.Sprintf("greeting.welcome")
	//
	//	// 格式化翻译
	//	message := formatter.Sprintf("user.welcome", "Alice")
	//
	//	// 标准格式化
	//	result := formatter.Sprintf("Count: %d", 42)
	Sprintf(format string, args ...any) string

	// Sprintln 使用当前语言环境格式化参数为字符串并添加换行符。
	//
	// 类似于 fmt.Sprintln，但支持本地化格式化。
	// 参数之间用空格分隔，末尾总是添加换行符。
	//
	// 参数：
	//   args: 要格式化的参数列表
	//
	// 返回：
	//   string: 格式化后带换行符的字符串
	//
	// 使用示例：
	//
	//	result := formatter.Sprintln("Hello", "World")
	//	// 输出: "Hello World\n" （根据语言环境可能调整格式）
	Sprintln(args ...any) string
}

// WriterFormatter 定义了写入式格式化接口。
//
// 此接口扩展了 Formatter，支持将格式化结果直接写入到 io.Writer。
// 这对于文件输出、网络响应、日志记录等场景特别有用。
//
// 性能优势：
// - 避免中间字符串分配
// - 直接写入目标 Writer
// - 支持流式处理
type WriterFormatter interface {
	// Fprint 使用当前语言环境格式化参数并写入到指定 Writer。
	//
	// 类似于 fmt.Fprint，但支持本地化格式化。
	// 所有参数按照它们的默认字符串表示进行连接并写入。
	//
	// 参数：
	//   w: 目标写入器（文件、网络连接、缓冲区等）
	//   args: 要写入的参数列表
	//
	// 返回：
	//   int: 写入的字节数
	//   error: 写入过程中遇到的错误
	//
	// 使用示例：
	//
	//	n, err := formatter.Fprint(os.Stdout, "Hello", " ", "World")
	//	if err != nil {
	//	    log.Fatal(err)
	//	}
	//	fmt.Printf("Wrote %d bytes\n", n)
	Fprint(w io.Writer, args ...any) (n int, err error)

	// Fprintf 使用当前语言环境格式化字符串并写入到指定 Writer。
	//
	// 类似于 fmt.Fprintf，但支持本地化翻译和格式化。
	// 支持翻译键查找和标准格式化动词。
	//
	// 参数：
	//   w: 目标写入器
	//   format: 格式字符串，可以包含翻译键
	//   args: 要插入格式字符串的参数
	//
	// 返回：
	//   int: 写入的字节数
	//   error: 写入过程中遇到的错误
	//
	// 使用示例：
	//
	//	n, err := formatter.Fprintf(os.Stdout, "user.welcome")
	//	if err != nil {
	//	    log.Fatal(err)
	//	}
	//
	//	// 标准格式化
	//	n, err = formatter.Fprintf(file, "Count: %d\n", 42)
	Fprintf(w io.Writer, format string, args ...any) (n int, err error)

	// Fprintln 使用当前语言环境格式化参数并写入到指定 Writer，添加换行符。
	//
	// 类似于 fmt.Fprintln，但支持本地化格式化。
	// 参数之间用空格分隔，末尾总是添加换行符。
	//
	// 参数：
	//   w: 目标写入器
	//   args: 要写入的参数列表
	//
	// 返回：
	//   int: 写入的字节数
	//   error: 写入过程中遇到的错误
	//
	// 使用示例：
	//
	//	n, err := formatter.Fprintln(os.Stdout, "Hello", "World")
	//	if err != nil {
	//	    log.Fatal(err)
	//	}
	Fprintln(w io.Writer, args ...any) (n int, err error)
}

// ConsoleFormatter 定义了控制台输出格式化接口。
//
// 此接口提供了直接输出到标准输出/错误的方法，
// 适合命令行工具、交互式应用程序和调试输出。
//
// 使用场景：
// - 命令行界面（CLI）工具
// - 调试和日志输出
// - 交互式应用程序
// - 开发和测试环境
type ConsoleFormatter interface {
	// Print 使用当前语言环境格式化参数并输出到标准输出。
	//
	// 类似于 fmt.Print，但支持本地化格式化。
	// 直接输出到 os.Stdout，适合命令行程序使用。
	//
	// 参数：
	//   args: 要输出的参数列表
	//
	// 返回：
	//   int: 写入的字节数
	//   error: 输出过程中遇到的错误
	//
	// 使用示例：
	//
	//	err := formatter.Print("Hello, ", "World!")
	//	if err != nil {
	//	    log.Fatal(err)
	//	}
	Print(args ...any) (n int, err error)

	// Printf 使用当前语言环境格式化字符串并输出到标准输出。
	//
	// 类似于 fmt.Printf，但支持本地化翻译和格式化。
	// 这是最常用的控制台输出方法，支持翻译和格式化。
	//
	// 参数：
	//   format: 格式字符串，可以包含翻译键
	//   args: 要插入格式字符串的参数
	//
	// 返回：
	//   int: 写入的字节数
	//   error: 输出过程中遇到的错误
	//
	// 使用示例：
	//
	//	// 输出翻译文本
	//	err := formatter.Printf("app.welcome")
	//
	//	// 输出格式化消息
	//	err = formatter.Printf("user.count: %d", userCount)
	Printf(format string, args ...any) (n int, err error)

	// Println 使用当前语言环境格式化参数并输出到标准输出，添加换行符。
	//
	// 类似于 fmt.Println，但支持本地化格式化。
	// 参数之间用空格分隔，末尾总是添加换行符。
	//
	// 参数：
	//   args: 要输出的参数列表
	//
	// 返回：
	//   int: 写入的字节数
	//   error: 输出过程中遇到的错误
	//
	// 使用示例：
	//
	//	err := formatter.Println("Hello", "World")
	//	if err != nil {
	//	    log.Fatal(err)
	//	}
	Println(args ...any) (n int, err error)
}

// Printer 完整的打印机接口，组合了所有功能。
//
// 此接口组合了 Localizer、Formatter、WriterFormatter 和 ConsoleFormatter，
// 提供了完整的翻译和格式化功能。所有的翻译实现都应该实现此接口。
//
// 接口组合优势：
// - 单一接口满足所有需求
// - 支持接口嵌套和组合
// - 便于模拟和测试
// - 保持向后兼容性
//
// 实现要求：
// - 必须线程安全
// - 支持缓存机制
// - 实现语言回退
// - 处理错误情况
type Printer interface {
	Localizer        // 提供语言环境访问
	Formatter        // 提供字符串格式化
	WriterFormatter  // 提供写入式格式化
	ConsoleFormatter // 提供控制台输出
}

// NewPrinter 创建由 fmt 驱动的打印器
func NewPrinter(locale Locale) Printer {
	return &simplePrinter{
		locale: locale,
	}
}

type simplePrinter struct {
	locale Locale
}

// Language 返回驱动支持的语言
func (d *simplePrinter) Locale() Locale {
	return d.locale
}

// Sprint 类似于 fmt.Sprint，但不添加空格分隔符
func (*simplePrinter) Sprint(args ...any) string {
	var result string
	for _, arg := range args {
		result += fmt.Sprint(arg)
	}
	return result
}

// Sprintf 类似于 fmt.Sprintf
func (*simplePrinter) Sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

// Sprintln 类似于 fmt.Sprintln
func (*simplePrinter) Sprintln(args ...any) string {
	return fmt.Sprintln(args...)
}

// Fprint 类似于 fmt.Fprint
func (*simplePrinter) Fprint(w io.Writer, args ...any) (n int, err error) {
	return fmt.Fprint(w, args...)
}

// Fprintf 类似于 fmt.Fprintf
func (*simplePrinter) Fprintf(w io.Writer, format string, args ...any) (n int, err error) {
	return fmt.Fprintf(w, format, args...)
}

// Fprintln 类似于 fmt.Fprintln
func (*simplePrinter) Fprintln(w io.Writer, args ...any) (n int, err error) {
	return fmt.Fprintln(w, args...)
}

// Print 类似于 fmt.Print
func (*simplePrinter) Print(args ...any) (n int, err error) {
	return fmt.Print(args...)
}

// Printf 类似于 fmt.Printf
func (*simplePrinter) Printf(format string, args ...any) (n int, err error) {
	return fmt.Printf(format, args...)
}

// Println 类似于 fmt.Println
func (*simplePrinter) Println(args ...any) (n int, err error) {
	return fmt.Println(args...)
}
