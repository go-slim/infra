// Package msg 提供全局翻译管理器的便利函数。
//
// 本包文件提供了全局默认 Manager 的访问接口，允许应用程序
// 在不需要显式传递 Manager 实例的情况下进行国际化和本地化操作。
// 这对于简单的应用程序或需要全局状态管理的场景特别有用。
//
// 全局管理器特点：
// 1. 懒加载初始化：只在第一次使用时创建
// 2. 线程安全：所有操作都是并发安全的
// 3. 可配置：支持运行时更换默认配置
// 4. 上下文感知：支持基于 context 的请求级本地化
// 5. 便利性：提供与 fmt 包类似的 API
//
// 使用场景：
// - 简单应用程序的快速国际化
// - 脚本工具的多语言支持
// - 需要全局翻译状态的应用
// - 原型开发和测试
//
// 注意事项：
// - 在微服务或多租户环境中，建议使用显式的 Manager 而非全局函数
// - 全局状态的修改会影响所有使用全局函数的代码
// - 在并发环境中，全局管理器是线程安全的
//
// 基本使用示例：
//
//	// 设置全局语言
//	msg.SetLocale(Chinese)
//
//	// 简单翻译
//	fmt.Println(msg.Sprintf("Hello, %s!", "World"))
//
//	// 使用上下文
//	ctx := msg.WithLocaleContext(context.Background(), French)
//	fmt.Println(msg.SprintfWithContext(ctx, "Welcome"))
//
// 高级配置示例：
//
//	factory := NewPrinterFactory(config)
//	manager := NewManager(ManagerConfig{
//	    Locale: English,
//	    Factory: factory,
//	})
//	msg.SetDefaultManager(manager)
package msg

import (
	"context"
	"io"
	"sync"
)

var (
	// defaultManager 全局默认 Manager 实例。
	// 使用懒加载模式，只在第一次需要时初始化。
	defaultManager *Manager

	// defaultOnce 确保 defaultManager 只初始化一次。
	// 在 SetDefaultManager 调用时可以被重置。
	defaultOnce sync.Once

	// defaultMutex 保护 defaultManager 的并发访问。
	// 使用读写锁优化读取性能。
	defaultMutex sync.RWMutex
)

// GetDefaultManager 获取全局默认的 Manager。
//
// 此函数使用懒加载模式，只在第一次调用时初始化全局 Manager。
// 如果没有显式设置，将创建一个使用英语作为默认语言的 Manager。
//
// 线程安全性：
// - 使用读写锁保护并发访问
// - 懒加载确保只初始化一次
// - 支持高并发读取场景
//
// 返回：
//
//	*Manager: 全局默认 Manager 实例
//
// 使用示例：
//
//	manager := msg.GetDefaultManager()
//	printer := manager.GetPrinter(Chinese)
//	message := printer.Sprintf("Hello")
func GetDefaultManager() *Manager {
	defaultMutex.RLock()
	if defaultManager != nil {
		manager := defaultManager
		defaultMutex.RUnlock()
		return manager
	}
	defaultMutex.RUnlock()

	// 使用 sync.Once 确保只初始化一次
	defaultOnce.Do(func() {
		defaultMutex.Lock()
		defaultManager = NewManager(ManagerConfig{
			Locale: English, // 默认使用英语
		})
		defaultMutex.Unlock()
	})

	defaultMutex.RLock()
	defer defaultMutex.RUnlock()
	return defaultManager
}

// SetDefaultManager 设置全局默认的 Manager。
//
// 此函数允许应用程序使用自定义的 Manager 替换默认实例。
// 这会影响所有后续使用全局函数的调用，包括已进行懒加载的场景。
//
// 注意事项：
// - 此操作是全局性的，会影响所有使用全局函数的代码
// - 在多线程环境中调用是安全的
// - 重置后，原有的懒加载状态会被清除
//
// 参数：
//
//	manager: 要设置为默认的 Manager 实例，可以为 nil
//
// 使用示例：
//
//	factory := NewPrinterFactory(config)
//	customManager := NewManager(ManagerConfig{
//	    Locale: French,
//	    Factory: factory,
//	})
//
//	// 设置全局管理器
//	msg.SetDefaultManager(customManager)
//
//	// 现在所有全局函数都将使用新的管理器
//	fmt.Println(msg.Sprintf("Hello")) // 将使用法语
func SetDefaultManager(manager *Manager) {
	defaultMutex.Lock()
	defer defaultMutex.Unlock()
	defaultManager = manager
	defaultOnce = sync.Once{} // 重置 Once，以便后续可以重新初始化
}

// SetLocale 设置全局默认语言。
//
// 此函数设置全局 Manager 的默认语言环境。
// 所有后续不指定语言的翻译操作都将使用这个语言环境。
//
// 参数：
//
//	locale: 要设置的默认语言环境
//
// 使用示例：
//
//	msg.SetLocale(Chinese)
//	fmt.Println(msg.Sprintf("Hello")) // 将使用中文翻译
//
//	msg.SetLocale(French)
//	fmt.Println(msg.Sprintf("Hello")) // 现在将使用法文翻译
func SetLocale(locale Locale) {
	GetDefaultManager().SetLocale(locale)
}

// GetLocale 获取全局默认语言。
//
// 返回当前全局 Manager 设置的默认语言环境。
//
// 返回：
//
//	Locale: 当前的默认语言环境
//
// 使用示例：
//
//	currentLocale := msg.GetLocale()
//	fmt.Printf("Current language: %s\n", currentLocale)
func GetLocale() Locale {
	return GetDefaultManager().GetLocale()
}

// SetPrinterFactory 设置全局默认的翻译工厂。
//
// 此函数设置全局 Manager 使用的翻译工厂。
// 翻译工厂负责根据语言环境创建对应的 Printer 实例。
//
// 参数：
//
//	factory: 要设置的翻译工厂
//
// 使用示例：
//
//	factory := NewPrinterFactory(config)
//	msg.SetPrinterFactory(factory)
//
//	// 现在所有翻译操作都使用新的工厂
//	printer := msg.GetPrinter()
func SetPrinterFactory(factory PrinterFactory) {
	GetDefaultManager().SetPrinterFactory(factory)
}

// GetPrinter 获取全局默认语言的 Printer。
//
// 返回使用全局默认语言环境配置的 Printer 实例。
// 该 Printer 可以用于后续的翻译和格式化操作。
//
// 返回：
//
//	Printer: 配置了默认语言环境的 Printer 实例
//
// 使用示例：
//
//	printer := msg.GetPrinter()
//	message := printer.Sprintf("Hello, World!")
//	fmt.Println(message)
func GetPrinter() Printer {
	return GetDefaultManager().GetPrinter()
}

// GetPrinterWithLocale 获取指定语言的 Printer。
//
// 返回使用指定语言环境配置的 Printer 实例，
// 不会影响全局默认语言设置。
//
// 参数：
//
//	locale: 要使用的语言环境
//
// 返回：
//
//	Printer: 配置了指定语言环境的 Printer 实例
//
// 使用示例：
//
//	printer := msg.GetPrinterWithLocale(French)
//	message := printer.Sprintf("Hello") // 将使用法语翻译
//
//	// 全局语言仍然是之前设置的
//	fmt.Println(msg.Sprintf("Hello")) // 可能使用其他语言
func GetPrinterWithLocale(locale Locale) Printer {
	return GetDefaultManager().GetPrinter(locale)
}

// Sprint 使用全局默认语言进行 Sprint 格式化。
//
// 相当于 fmt.Sprint 的国际化版本，使用全局默认语言环境。
// 所有参数都将按照它们的默认字符串表示进行连接。
//
// 参数：
//
//	args: 要格式化的参数列表
//
// 返回：
//
//	string: 格式化后的字符串
//
// 使用示例：
//
//	msg.SetLocale(Chinese)
//	result := msg.Sprint("Hello", 42, true)
//	// 可能会根据语言设置调整格式
func Sprint(args ...any) string {
	return GetDefaultManager().Sprint(args...)
}

// Sprintf 使用全局默认语言进行 Sprintf 格式化。
//
// 这是最常用的全局翻译函数，相当于 fmt.Sprintf 的国际化版本。
// 使用全局 Manager 的默认语言环境进行翻译和格式化。
//
// 参数：
//
//	format: 格式字符串，可以包含翻译键和格式化动词
//	args: 要插入格式字符串的参数
//
// 返回：
//
//	string: 翻译和格式化后的字符串
//
// 使用示例：
//
//	// 基本翻译
//	result := msg.Sprintf("Hello, %s!", "World")
//
//	// 在设置语言后
//	msg.SetLocale(Chinese)
//	result := msg.Sprintf("Welcome") // 将返回中文翻译
func Sprintf(format string, args ...any) string {
	return GetDefaultManager().Sprintf(format, args...)
}

// Sprintln 使用全局默认语言进行 Sprintln 格式化
func Sprintln(args ...any) string {
	return GetDefaultManager().Sprintln(args...)
}

// Print 使用全局默认语言进行 Print 格式化
func Print(args ...any) (n int, err error) {
	return GetDefaultManager().Print(args...)
}

// Printf 使用全局默认语言进行 Printf 格式化
func Printf(format string, args ...any) (n int, err error) {
	return GetDefaultManager().Printf(format, args...)
}

// Println 使用全局默认语言进行 Println 格式化
func Println(args ...any) (n int, err error) {
	return GetDefaultManager().Println(args...)
}

// Fprint 使用全局默认语言进行 Fprint 格式化
func Fprint(w io.Writer, args ...any) (n int, err error) {
	return GetDefaultManager().Fprint(w, args...)
}

// Fprintf 使用全局默认语言进行 Fprintf 格式化
func Fprintf(w io.Writer, format string, args ...any) (n int, err error) {
	return GetDefaultManager().Fprintf(w, format, args...)
}

// Fprintln 使用全局默认语言进行 Fprintln 格式化
func Fprintln(w io.Writer, args ...any) (n int, err error) {
	return GetDefaultManager().Fprintln(w, args...)
}

// GetPrinterWithContext 从上下文获取 Printer
func GetPrinterWithContext(ctx context.Context) Printer {
	return GetDefaultManager().GetPrinterWithContext(ctx)
}

// SprintWithContext 使用上下文语言进行 Sprint 格式化
func SprintWithContext(ctx context.Context, args ...any) string {
	return GetDefaultManager().SprintWithContext(ctx, args...)
}

// SprintfWithContext 使用上下文语言进行 Sprintf 格式化
func SprintfWithContext(ctx context.Context, format string, args ...any) string {
	return GetDefaultManager().SprintfWithContext(ctx, format, args...)
}

// SprintlnWithContext 使用上下文语言进行 Sprintln 格式化
func SprintlnWithContext(ctx context.Context, args ...any) string {
	return GetDefaultManager().SprintlnWithContext(ctx, args...)
}

// PrintWithContext 使用上下文语言进行 Print 格式化
func PrintWithContext(ctx context.Context, args ...any) (n int, err error) {
	return GetDefaultManager().PrintWithContext(ctx, args...)
}

// PrintfWithContext 使用上下文语言进行 Printf 格式化
func PrintfWithContext(ctx context.Context, format string, args ...any) (n int, err error) {
	return GetDefaultManager().PrintfWithContext(ctx, format, args...)
}

// PrintlnWithContext 使用上下文语言进行 Println 格式化
func PrintlnWithContext(ctx context.Context, args ...any) (n int, err error) {
	return GetDefaultManager().PrintlnWithContext(ctx, args...)
}

// FprintWithContext 使用上下文语言进行 Fprint 格式化
func FprintWithContext(ctx context.Context, w io.Writer, args ...any) (n int, err error) {
	return GetDefaultManager().FprintWithContext(ctx, w, args...)
}

// FprintfWithContext 使用上下文语言进行 Fprintf 格式化
func FprintfWithContext(ctx context.Context, w io.Writer, format string, args ...any) (n int, err error) {
	return GetDefaultManager().FprintfWithContext(ctx, w, format, args...)
}

// FprintlnWithContext 使用上下文语言进行 Fprintln 格式化
func FprintlnWithContext(ctx context.Context, w io.Writer, args ...any) (n int, err error) {
	return GetDefaultManager().FprintlnWithContext(ctx, w, args...)
}

// WithLocale 使用指定语言执行函数
func WithLocale(locale Locale, fn func(Printer)) {
	GetDefaultManager().WithLocale(locale, fn)
}

// WithContext 使用上下文信息执行函数
func WithContext(ctx context.Context, fn func(Printer)) {
	GetDefaultManager().WithContext(ctx, fn)
}
