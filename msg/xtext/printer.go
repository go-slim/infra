package xtext

import (
	"fmt"
	"io"

	"go-slim.dev/infra/msg"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Printer 基于 golang.org/x/text 的 msg.Printer 实现。
//
// Printer 是翻译系统的核心组件，负责将翻译键转换为本地化文本。
// 它包装了 golang.org/x/text/message.Printer，提供了完整的格式化功能。
//
// 主要功能：
// 1. 基于 BCP 47 语言标签的本地化
// 2. 支持多种输出格式（字符串、文件、控制台）
// 3. 完整的格式化方法集合（Print, Printf, Sprint, Sprintf 等）
// 4. 与 msg 包接口的完全兼容
//
// 使用示例：
//
//	printer, err := xtext.NewPrinter(msg.Locale("zh-CN"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 简单翻译
//	fmt.Println(printer.Sprintf("Hello, %s!", "World"))
//
//	// 输出到文件
//	err = printer.Fprintf(os.Stdout, "Translated: %s\n", "message")
//
//	// 控制台输出
//	printer.Printf("Message: %s\n", "content")
type Printer struct {
	printer *message.Printer // 底层的 golang.org/x/text 打印机
	locale  msg.Locale       // 关联的语言环境
}

// NewPrinter 创建新的 xtext 打印机。
//
// 该函数创建一个基于指定语言环境的 Printer 实例。
// 语言标签将使用 BCP 47 标准格式进行解析。
//
// 参数：
//
//	locale: 目标语言环境，如 "zh-CN", "en-US", "fr-FR"
//	opts: 可选的 message.Option 配置参数
//
// 返回：
//   - msg.Printer: 新创建的打印机实例
//   - error: 如果语言标签解析失败，返回错误
//
// 示例：
//
//	printer, err := NewPrinter(msg.Locale("zh-CN"))
//	if err != nil {
//	    return nil, fmt.Errorf("failed to create printer: %w", err)
//	}
//	return printer, nil
func NewPrinter(locale msg.Locale, opts ...message.Option) (msg.Printer, error) {
	// 将 msg.Locale 转换为 golang.org/x/text/language.Tag
	tag, err := language.All.Parse(locale.String())
	if err != nil {
		// 语言标签解析失败，使用英语作为后备，但仍返回错误信息
		printer := &Printer{
			printer: message.NewPrinter(language.English, opts...),
			locale:  locale, // 保持原始 locale 信息
		}
		return printer, fmt.Errorf("failed to parse locale '%s': %w", locale, err)
	}

	return &Printer{
		printer: message.NewPrinter(tag, opts...),
		locale:  locale,
	}, nil
}

// Locale 实现 msg.Localizer 接口。
//
// 返回与此 Printer 关联的语言环境。
// 这个信息可以用于日志记录、调试或需要知道当前语言环境的场景。
//
// 返回：
//
//	msg.Locale: 关联的语言环境
func (p *Printer) Locale() msg.Locale {
	return p.locale
}

// Sprint 实现 msg.Formatter 接口。
//
// 使用默认格式将参数格式化为字符串，但不添加空格分隔符。
// 所有参数都将按照它们的默认字符串表示进行直接连接。
//
// 参数：
//
//	args: 要格式化的参数列表
//
// 返回：
//
//	string: 格式化后的字符串
//
// 示例：
//
//	result := printer.Sprint("Hello", 42, true)
//	// 输出: "Hello42true"
func (p *Printer) Sprint(args ...any) string {
	var result string
	for _, arg := range args {
		result += fmt.Sprint(arg)
	}
	return result
}

// Sprintf 实现 msg.Formatter 接口。
//
// 使用格式字符串将参数格式化为本地化文本，相当于 fmt.Sprintf。
// 支持所有标准的 Go 格式化动词，并且会应用当前语言环境的翻译规则。
//
// 参数：
//
//	format: 格式字符串，可以包含翻译键和格式化动词
//	args: 要插入格式字符串的参数
//
// 返回：
//
//	string: 本地化和格式化后的字符串
//
// 示例：
//
//	result := printer.Sprintf("Hello, %s!", "World")
//	// 如果有翻译，输出本地化结果，否则输出: "Hello, World!"
func (p *Printer) Sprintf(format string, args ...any) string {
	return p.printer.Sprintf(format, args...)
}

// Sprintln 实现 msg.Formatter 接口。
//
// 使用默认格式将参数格式化为字符串，并添加换行符，相当于 fmt.Sprintln。
// 参数之间用空格分隔，末尾总是添加换行符。
//
// 参数：
//
//	args: 要格式化的参数列表
//
// 返回：
//
//	string: 格式化后带换行符的字符串
//
// 示例：
//
//	result := printer.Sprintln("Hello", "World")
//	// 输出: "Hello World\n"
func (p *Printer) Sprintln(args ...any) string {
	return p.printer.Sprintln(args...)
}

// Fprint 实现 msg.WriterFormatter 接口。
//
// 使用默认格式将参数写入指定的写入器，相当于 fmt.Fprint。
// 所有参数都将按照它们的默认字符串表示进行连接并写入。
//
// 参数：
//
//	w: 目标写入器（如 os.Stdout, os.Stderr, 文件等）
//	args: 要写入的参数列表
//
// 返回：
//
//	int: 写入的字节数
//	error: 写入过程中遇到的错误
//
// 示例：
//
//	n, err := printer.Fprint(os.Stdout, "Hello", " ", "World")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (p *Printer) Fprint(w io.Writer, args ...any) (n int, err error) {
	return p.printer.Fprint(w, args...)
}

// Fprintf 实现 msg.WriterFormatter 接口。
//
// 使用格式字符串将参数写入指定的写入器，相当于 fmt.Fprintf。
// 支持所有标准的 Go 格式化动词，并且会应用当前语言环境的翻译规则。
//
// 参数：
//
//	w: 目标写入器（如 os.Stdout, os.Stderr, 文件等）
//	format: 格式字符串，可以包含翻译键和格式化动词
//	args: 要插入格式字符串的参数
//
// 返回：
//
//	int: 写入的字节数
//	error: 写入过程中遇到的错误
//
// 示例：
//
//	n, err := printer.Fprintf(os.Stdout, "Hello, %s!\n", "World")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (p *Printer) Fprintf(w io.Writer, format string, args ...any) (n int, err error) {
	return p.printer.Fprintf(w, format, args...)
}

// Fprintln 实现 msg.WriterFormatter 接口。
//
// 使用默认格式将参数写入指定的写入器，并添加换行符，相当于 fmt.Fprintln。
// 参数之间用空格分隔，末尾总是添加换行符。
//
// 参数：
//
//	w: 目标写入器（如 os.Stdout, os.Stderr, 文件等）
//	args: 要写入的参数列表
//
// 返回：
//
//	int: 写入的字节数
//	error: 写入过程中遇到的错误
//
// 示例：
//
//	n, err := printer.Fprintln(os.Stdout, "Hello", "World")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (p *Printer) Fprintln(w io.Writer, args ...any) (n int, err error) {
	return p.printer.Fprintln(w, args...)
}

// Print 实现 msg.ConsoleFormatter 接口。
//
// 使用默认格式将参数写入标准输出，相当于 fmt.Print。
// 所有参数都将按照它们的默认字符串表示进行连接并输出到控制台。
//
// 参数：
//
//	args: 要输出的参数列表
//
// 返回：
//
//	int: 写入的字节数
//	error: 写入过程中遇到的错误
//
// 示例：
//
//	err := printer.Print("Hello", " ", "World")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (p *Printer) Print(args ...any) (n int, err error) {
	return p.printer.Print(args...)
}

// Printf 实现 msg.ConsoleFormatter 接口。
//
// 使用格式字符串将参数写入标准输出，相当于 fmt.Printf。
// 支持所有标准的 Go 格式化动词，并且会应用当前语言环境的翻译规则。
// 这是最常用的格式化输出方法。
//
// 参数：
//
//	format: 格式字符串，可以包含翻译键和格式化动词
//	args: 要插入格式字符串的参数
//
// 返回：
//
//	int: 写入的字节数
//	error: 写入过程中遇到的错误
//
// 示例：
//
//	err := printer.Printf("Hello, %s!\n", "World")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 使用翻译键
//	err = printer.Printf("greeting.welcome\n")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (p *Printer) Printf(format string, args ...any) (n int, err error) {
	return p.printer.Printf(format, args...)
}

// Println 实现 msg.ConsoleFormatter 接口。
//
// 使用默认格式将参数写入标准输出，并添加换行符，相当于 fmt.Println。
// 参数之间用空格分隔，末尾总是添加换行符。
//
// 参数：
//
//	args: 要输出的参数列表
//
// 返回：
//
//	int: 写入的字节数
//	error: 写入过程中遇到的错误
//
// 示例：
//
//	err := printer.Println("Hello", "World")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (p *Printer) Println(args ...any) (n int, err error) {
	return p.printer.Println(args...)
}
