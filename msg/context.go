// Package msg 提供了基于 Go context 的国际化和本地化支持。
//
// 本包允许应用程序在上下文中传播语言环境和翻译工厂信息，
// 实现请求级别的本地化。这对于 Web 应用程序、微服务架构
// 等需要根据用户偏好动态切换语言的场景特别有用。
//
// 主要功能：
// 1. 上下文传播：在 context.Context 中传递语言环境和工厂信息
// 2. 请求级本地化：支持不同请求使用不同的语言环境
// 3. 工厂模式：通过 PrinterFactory 接口支持多种翻译实现
// 4. 线程安全：所有操作都是并发安全的
// 5. 回退机制：当上下文中没有指定语言时，提供回退选项
//
// 设计理念：
// - 零依赖：仅使用标准库的 context 包
// - 类型安全：使用强类型的 Locale 和 PrinterFactory 接口
// - 链式调用：支持函数式编程风格的上下文构建
// - 性能优化：使用结构体键避免冲突，减少内存分配
//
// 使用示例：
//
//	// 创建带语言环境的上下文
//	ctx := WithLocaleContext(context.Background(), Chinese)
//
//	// 添加翻译工厂
//	ctx = WithPrinterFactoryContext(ctx, factory)
//
//	// 一次性添加语言和工厂
//	ctx = WithLocaleAndPrinterFactoryContext(ctx, English, factory)
//
//	// 从上下文中获取语言信息
//	if locale, ok := GetLocaleFromContext(ctx); ok {
//	    fmt.Printf("Current locale: %s\n", locale)
//	}
//
//	// 从上下文中获取翻译工厂
//	if factory, ok := GetPrinterFactoryFromContext(ctx); ok {
//	    printer := factory.CreatePrinter(locale)
//	    result := printer.Sprintf("Hello, %s!", "World")
//	}
package msg

import "context"

// contextLocaleKey 是上下文中存储语言信息的键。
//
// 使用不导出的结构体类型作为键，避免与其他包在 context 中
// 的键发生冲突。这是 Go 中 context 键的最佳实践。
var contextLocaleKey = struct{ name string }{name: "msg-locale"}

// contextPrinterFactoryKey 是上下文中存储翻译工厂的键。
//
// 使用不导出的结构体类型作为键，确保键的唯一性。
// 这种设计防止了不同包之间的键冲突。
var contextPrinterFactoryKey = struct{ name string }{name: "msg-printer-factory"}

// nilFactorySentinel 用于在上下文中明确表示存储了 nil 工厂
type nilFactorySentinel struct{}

// WithLocaleContext 将语言信息添加到上下文中。
//
// 此函数创建一个新的上下文，其中包含指定的语言环境。
// 原始上下文不会被修改，符合 Go context 的不可变特性。
//
// 参数：
//
//	ctx: 原始上下文，通常为 context.Background() 或现有的请求上下文
//	locale: 要设置的语言环境，如 Chinese, English, French 等
//
// 返回：
//
//	context.Context: 包含语言信息的新上下文
//
// 使用示例：
//
//	func handleRequest(w http.ResponseWriter, r *http.Request) {
//	    // 从请求头或用户设置中获取语言偏好
//	    locale := determineLocaleFromRequest(r)
//
//	    // 创建带语言信息的上下文
//	    ctx := WithLocaleContext(r.Context(), locale)
//
//	    // 将上下文传递给后续处理函数
//	    processRequest(ctx)
//	}
func WithLocaleContext(ctx context.Context, locale Locale) context.Context {
	return context.WithValue(ctx, contextLocaleKey, locale)
}

// WithPrinterFactoryContext 将翻译工厂信息添加到上下文中。
//
// 此函数创建一个新的上下文，其中包含指定的翻译工厂。
// 翻译工厂负责根据语言环境创建对应的 Printer 实例。
//
// 参数：
//
//	ctx: 原始上下文
//	factory: 翻译工厂实例，实现 PrinterFactory 接口
//
// 返回：
//
//	context.Context: 包含翻译工厂信息的新上下文
//
// 使用示例：
//
//	factory := NewPrinterFactory(config)
//	ctx := WithPrinterFactoryContext(context.Background(), factory)
//
//	// 在后续代码中可以从上下文获取工厂
//	if f, ok := GetPrinterFactoryFromContext(ctx); ok {
//	    printer := f.CreatePrinter(English)
//	    message := printer.Sprintf("Welcome!")
//	}
func WithPrinterFactoryContext(ctx context.Context, factory PrinterFactory) context.Context {
	if factory == nil {
		// Store sentinel value to distinguish explicit nil from missing key
		return context.WithValue(ctx, contextPrinterFactoryKey, nilFactorySentinel{})
	}
	return context.WithValue(ctx, contextPrinterFactoryKey, factory)
}

// WithLocaleAndPrinterFactoryContext 将语言和翻译工厂信息同时添加到上下文中。
//
// 这是一个便利函数，一次性将语言环境和翻译工厂都添加到上下文中。
// 比分别调用 WithLocaleContext 和 WithPrinterFactoryContext 更高效。
//
// 参数：
//
//	ctx: 原始上下文
//	locale: 要设置的语言环境
//	factory: 翻译工厂实例
//
// 返回：
//
//	context.Context: 同时包含语言和翻译工厂信息的新上下文
//
// 使用示例：
//
//	factory := NewPrinterFactory(config)
//	ctx := WithLocaleAndPrinterFactoryContext(
//	    context.Background(),
//	    Chinese,
//	    factory,
//	)
//
//	// 现在上下文中同时包含了语言和工厂信息
//	locale, _ := GetLocaleFromContext(ctx)
//	factory, _ := GetPrinterFactoryFromContext(ctx)
//	printer := factory.CreatePrinter(locale)
func WithLocaleAndPrinterFactoryContext(ctx context.Context, locale Locale, factory PrinterFactory) context.Context {
	ctx = context.WithValue(ctx, contextLocaleKey, locale)
	ctx = context.WithValue(ctx, contextPrinterFactoryKey, factory)
	return ctx
}

// GetLocaleFromContext 从上下文中获取语言信息。
//
// 此函数从上下文中提取之前设置的语言环境。支持可选的回退语言，
// 当上下文中没有找到语言信息时，可以使用提供的回退语言。
//
// 参数：
//
//	ctx: 要查询的上下文
//	fallback: 可选的回退语言列表，按优先级顺序使用
//
// 返回：
//
//	Locale: 找到的语言环境或回退语言
//	bool: 是否从上下文中成功获取到语言信息（不包括回退）
//
// 使用示例：
//
//	// 基本用法
//	if locale, ok := GetLocaleFromContext(ctx); ok {
//	    fmt.Printf("Found locale in context: %s\n", locale)
//	} else {
//	    fmt.Println("No locale found in context")
//	}
//
//	// 使用回退语言
//	locale, ok := GetLocaleFromContext(ctx, English)
//	// 如果上下文中没有语言信息，将返回 English 作为回退
//	// ok 仍然为 false，表示是从回退获取的
func GetLocaleFromContext(ctx context.Context, fallback ...Locale) (Locale, bool) {
	if locale, ok := ctx.Value(contextLocaleKey).(Locale); ok {
		return locale, true
	}

	// 使用回退语言，按提供顺序检查
	for _, fb := range fallback {
		if fb != "" {
			return fb, true
		}
	}
	return "", false
}

// GetPrinterFactoryFromContext 从上下文中获取翻译工厂信息。
//
// 此函数从上下文中提取之前设置的翻译工厂。支持可选的回退工厂，
// 当上下文中没有找到工厂信息时，可以使用提供的回退工厂。
//
// 参数：
//
//	ctx: 要查询的上下文
//	fallback: 可选的回退翻译工厂列表，按优先级顺序使用
//
// 返回：
//
//	PrinterFactory: 找到的翻译工厂或回退工厂
//	bool: 是否从上下文中成功获取到工厂信息（不包括回退）
//
// 使用示例：
//
//	// 基本用法
//	if factory, ok := GetPrinterFactoryFromContext(ctx); ok {
//	    printer := factory.CreatePrinter(Chinese)
//	    message := printer.Sprintf("Hello")
//	} else {
//	    log.Println("No factory found in context")
//	}
//
//	// 使用回退工厂
//	defaultFactory := NewDefaultFactory()
//	factory, ok := GetPrinterFactoryFromContext(ctx, defaultFactory)
//	// 如果上下文中没有工厂，将使用 defaultFactory 作为回退
//	// ok 仍然为 false，表示是从回退获取的
func GetPrinterFactoryFromContext(ctx context.Context, fallback ...PrinterFactory) (PrinterFactory, bool) {
	// Check if context has the key at all
	if val := ctx.Value(contextPrinterFactoryKey); val != nil {
		// Handle sentinel value for explicit nil
		if _, ok := val.(nilFactorySentinel); ok {
			return nil, true
		}
		// Regular factory
		if factory, ok := val.(PrinterFactory); ok {
			return factory, true
		}
	}

	// 使用回退工厂，按提供顺序检查
	for _, fb := range fallback {
		if fb != nil {
			return fb, true
		}
	}
	return nil, false
}
