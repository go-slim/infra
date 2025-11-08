package msg

import (
	"cmp"
	"context"
	"io"
	"sync"
)

// LogFunc 定义日志函数类型
//
// 这个类型用于定义日志记录行为，允许用户自定义日志格式和输出目标。
// 用户可以在输出时自己添加 [LEVEL] 前缀，如：[INFO], [ERROR], [DEBUG] 等。
//
// 示例：
//
//	func(msg string) {
//	    log.Printf("[INFO] %s", msg)
//	}
type LogFunc func(string)

// Manager 管理多个语言的 Printer 实例，提供统一的接口来创建、获取和切换不同语言的 Printer。
//
// Manager 是国际化的核心组件，负责：
// 1. 管理当前语言设置
// 2. 缓存和重用 Printer 实例
// 3. 提供语言降级匹配机制
// 4. 支持上下文相关的语言切换
//
// 线程安全：所有公共方法都是线程安全的，可以在并发环境中使用。
//
// 使用示例：
//
//	// 创建管理器
//	manager := msg.NewManager(msg.ManagerConfig{
//	    Locale: msg.English,
//	    Factory: customFactory,
//	})
//
//	// 获取当前语言的打印机
//	printer := manager.GetPrinter()
//
//	// 获取特定语言的打印机
//	printer := manager.GetPrinter(msg.Chinese)
type Manager struct {
	mu      sync.RWMutex          // 读写锁，保护并发访问
	locale  Locale                // 当前语言设置
	logFunc LogFunc               // 日志函数
	factory *simplePrinterFactory // 内部打印机工厂
}

// ManagerConfig 管理器配置选项，用于创建 Manager 实例。
//
// 通过这个结构体可以配置管理器的行为：
// - 日志记录方式
// - 使用的打印机工厂
// - 默认语言设置
type ManagerConfig struct {
	// LogFunc 日志函数，如果为 nil 则不记录日志
	// 可以自定义日志格式，比如添加时间戳、日志级别等
	LogFunc LogFunc

	// Factory 驱动工厂，如果为 nil 则使用默认的 FmtPrinterFactory
	// 支持自定义的打印机实现，比如支持 i18n 的打印机
	Factory PrinterFactory

	// Locale 默认语言，也作为不支持语言时的回退语言
	// 如果为空则使用英语 "en"
	// 用于在没有指定语言时提供默认的本地化支持
	Locale Locale
}

// NewManager 创建新的管理器实例。
//
// 根据 ManagerConfig 配置创建 Manager：
// 1. 设置默认语言（优先使用配置，回退到英语）
// 2. 初始化打印机工厂（默认或自定义）
// 3. 设置回退语言
//
// 参数 config: 管理器配置
// 返回: 初始化完成的 Manager 实例
func NewManager(config ManagerConfig) *Manager {
	var factory *simplePrinterFactory

	if config.Factory == nil {
		// 如果没有提供工厂，创建默认的 simplePrinterFactory
		factory = NewPrinterFactory().(*simplePrinterFactory)
	} else {
		// 如果提供了外部工厂，包装它
		factory = NewPrinterFactory(config.Factory).(*simplePrinterFactory)
	}

	m := &Manager{
		locale:  cmp.Or(config.Locale, English), // 使用配置的语言或默认英语
		logFunc: config.LogFunc,
		factory: factory,
	}
	m.factory.SetFallbackLocale(m.locale)
	return m
}

// resolveLocale 将通用语言解析为更具体的默认语言环境
//
// 此方法处理常见的语言映射，将通用语言代码转换为更具体的默认形式：
// - zh (中文) -> zh-Hans-CN (简体中文-中国大陆)
//
// 参数 locale: 要解析的语言环境
// 返回: 解析后的具体语言环境
func (m *Manager) resolveLocale(locale Locale) Locale {
	// 目前只处理中文的特殊映射
	// zh 是通用中文，应该映射到最常用的 zh-Hans-CN
	if locale == Chinese {
		return Locale("zh-Hans-CN")
	}

	// 其他语言暂时保持原样，可以根据需要添加更多映射
	return locale
}

// log 记录日志的内部方法
func (m *Manager) log(message string) {
	if m.logFunc != nil {
		m.logFunc(message)
	}
}

// SetLocale 设置当前语言环境
func (m *Manager) SetLocale(locale Locale) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.log("[INFO] Setting locale: " + string(locale))
	m.locale = locale
	m.factory.SetFallbackLocale(locale)
}

// GetLocale 获取当前语言环境
func (m *Manager) GetLocale() Locale {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.locale
}

// SupportedLocales 获取驱动工厂支持的语言环境列表
func (m *Manager) SupportedLocales() []Locale {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回驱动工厂支持的语言列表
	supportedLocales := m.factory.SupportedLocales()

	// 如果驱动工厂返回 nil 或空列表，表示支持任何语言
	if len(supportedLocales) == 0 {
		return []Locale{}
	}

	// 返回驱动工厂明确支持的语言列表
	return supportedLocales
}

// SupportsLocale 检查当前驱动工厂是否支持指定语言
func (m *Manager) SupportsLocale(locale Locale) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.factory.SupportsLocale(locale)
}

// GetPrinter 获取指定语言的 Printer
// 如果没有提供语言参数，返回当前语言的 Printer
func (m *Manager) GetPrinter(locale ...Locale) Printer {
	var targetLocale Locale

	if len(locale) == 0 {
		// 没有提供语言参数，使用当前语言
		m.mu.RLock()
		targetLocale = m.locale
		m.mu.RUnlock()
	} else {
		// 使用提供的语言参数
		targetLocale = locale[0]
	}

	// 直接通过工厂创建，工厂内部会处理缓存
	m.mu.RLock()
	factory := m.factory
	m.mu.RUnlock()

	printer, err := factory.CreatePrinter(targetLocale)
	if err != nil {
		// 如果创建失败，使用简单的 Printer 作为后备
		m.log("[ERROR] failed to create printer for locale " + string(targetLocale) + ": " + err.Error() + ", using fallback fmt printer")
		return NewPrinter(targetLocale)
	}

	return printer
}

// SetPrinterFactory 设置新的驱动工厂
func (m *Manager) SetPrinterFactory(factory PrinterFactory) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.log("[INFO] Setting new PrinterFactory")
	m.factory.SetCustom(factory)
	m.factory.SetFallbackLocale(m.locale)
}

// GetPrinterFactory 获取当前的驱动工厂
func (m *Manager) GetPrinterFactory() PrinterFactory {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.factory
}

// findSuitableFactory 查找支持指定语言的工厂
// 按优先级检查：contextFactory -> managerFactory -> fallback
func (m *Manager) findSuitableFactory(locale Locale, contextFactory PrinterFactory, usingContext bool) (PrinterFactory, Locale) {
	// 检查当前工厂是否支持语言
	if checkFactorySupport(contextFactory, locale) {
		if usingContext {
			m.log("[DEBUG] Context factory supports requested locale: " + string(locale))
		} else {
			m.log("[DEBUG] Manager factory supports requested locale: " + string(locale))
		}
		return contextFactory, locale
	}

	// 如果是上下文工厂且不支持，尝试 Manager 工厂
	if usingContext {
		m.mu.RLock()
		managerFactory := m.factory
		m.mu.RUnlock()

		if checkFactorySupport(managerFactory, locale) {
			m.log("[INFO] Manager factory supports requested locale: " + string(locale))
			return managerFactory, locale
		}
	}

	// 都不支持，使用 fallback 语言
	m.mu.RLock()
	defaultLocale := m.locale
	m.mu.RUnlock()
	m.log("[WARN] Neither context nor Manager factory supports locale, using fallback: " + string(locale) + " -> " + string(defaultLocale))
	return contextFactory, defaultLocale
}

// checkFactorySupport 检查工厂是否支持指定语言
func checkFactorySupport(factory PrinterFactory, locale Locale) bool {
	if factory == nil {
		return false
	}

	supportedLocales := factory.SupportedLocales()
	// nil LocaleSet means unlimited support
	if supportedLocales == nil {
		return true
	}
	if len(supportedLocales) == 0 {
		return false
	}

	for _, supportedLocale := range supportedLocales {
		if supportedLocale.Contains(locale) || locale.Contains(supportedLocale) {
			return true
		}
	}
	return false
}

// 便利方法：直接使用当前语言进行格式化

// Sprint 使用当前语言进行 Sprint 格式化
func (m *Manager) Sprint(args ...any) string {
	return m.GetPrinter().Sprint(args...)
}

// Sprintf 使用当前语言进行 Sprintf 格式化
func (m *Manager) Sprintf(format string, args ...any) string {
	return m.GetPrinter().Sprintf(format, args...)
}

// Sprintln 使用当前语言进行 Sprintln 格式化
func (m *Manager) Sprintln(args ...any) string {
	return m.GetPrinter().Sprintln(args...)
}

// Print 使用当前语言进行 Print 格式化
func (m *Manager) Print(args ...any) (n int, err error) {
	return m.GetPrinter().Print(args...)
}

// Printf 使用当前语言进行 Printf 格式化
func (m *Manager) Printf(format string, args ...any) (n int, err error) {
	return m.GetPrinter().Printf(format, args...)
}

// Println 使用当前语言进行 Println 格式化
func (m *Manager) Println(args ...any) (n int, err error) {
	return m.GetPrinter().Println(args...)
}

// Fprint 使用当前语言进行 Fprint 格式化
func (m *Manager) Fprint(w io.Writer, args ...any) (n int, err error) {
	return m.GetPrinter().Fprint(w, args...)
}

// Fprintf 使用当前语言进行 Fprintf 格式化
func (m *Manager) Fprintf(w io.Writer, format string, args ...any) (n int, err error) {
	return m.GetPrinter().Fprintf(w, format, args...)
}

// Fprintln 使用当前语言进行 Fprintln 格式化
func (m *Manager) Fprintln(w io.Writer, args ...any) (n int, err error) {
	return m.GetPrinter().Fprintln(w, args...)
}

// WithLocale 使用指定语言执行函数，不改变 Manager 的当前语言状态
func (m *Manager) WithLocale(locale Locale, fn func(Printer)) {
	resolvedLocale := m.resolveLocale(locale)
	fn(m.GetPrinter(resolvedLocale))
}

// LocaleFromContext 从上下文中获取语言
// 如果上下文中没有语言信息，返回当前语言
func (m *Manager) LocaleFromContext(ctx context.Context) Locale {
	// 尝试从上下文中获取语言
	if locale, ok := GetLocaleFromContext(ctx); ok {
		return locale
	}

	// 如果上下文中没有语言信息，使用当前语言
	return m.GetLocale()
}

// GetPrinterWithContext 获取基于上下文语言和驱动的临时 Printer
// 优先使用上下文中的 PrinterFactory，如果不支持则尝试 Manager 的默认 PrinterFactory，
// 最后才使用 fallback 策略
func (m *Manager) GetPrinterWithContext(ctx context.Context) Printer {
	locale := m.LocaleFromContext(ctx)
	m.log("[INFO] Creating printer from context for locale: " + string(locale))

	// 获取上下文中的 PrinterFactory
	var factory PrinterFactory
	usingContext := false
	if ctxFactory, ok := GetPrinterFactoryFromContext(ctx); ok {
		factory = ctxFactory
		usingContext = true
		m.log("[DEBUG] Using PrinterFactory from context")
	} else {
		m.mu.RLock()
		factory = m.factory
		m.mu.RUnlock()
	}

	// 查找合适的工厂和目标语言
	finalFactory, targetLocale := m.findSuitableFactory(locale, factory, usingContext)

	// 创建临时 Printer（不缓存到 Manager 中）
	printer, err := finalFactory.CreatePrinter(targetLocale)
	if err != nil {
		m.log("[ERROR] failed to create printer for locale " + string(targetLocale) + ": " + err.Error() + ", using fallback fmt printer")
		printer = NewPrinter(targetLocale)
	}

	return printer
}

// WithContext 使用上下文中的语言信息执行函数，不改变 Manager 的当前语言状态
// 如果上下文中没有语言信息，使用当前语言
func (m *Manager) WithContext(ctx context.Context, fn func(Printer)) {
	fn(m.GetPrinterWithContext(ctx))
}

// SprintWithContext 使用上下文语言进行 Sprint 格式化
func (m *Manager) SprintWithContext(ctx context.Context, args ...any) string {
	return m.GetPrinterWithContext(ctx).Sprint(args...)
}

// SprintfWithContext 使用上下文语言进行 Sprintf 格式化
func (m *Manager) SprintfWithContext(ctx context.Context, format string, args ...any) string {
	return m.GetPrinterWithContext(ctx).Sprintf(format, args...)
}

// SprintlnWithContext 使用上下文语言进行 Sprintln 格式化
func (m *Manager) SprintlnWithContext(ctx context.Context, args ...any) string {
	return m.GetPrinterWithContext(ctx).Sprintln(args...)
}

// PrintWithContext 使用上下文语言进行 Print 格式化
func (m *Manager) PrintWithContext(ctx context.Context, args ...any) (n int, err error) {
	return m.GetPrinterWithContext(ctx).Print(args...)
}

// PrintfWithContext 使用上下文语言进行 Printf 格式化
func (m *Manager) PrintfWithContext(ctx context.Context, format string, args ...any) (n int, err error) {
	return m.GetPrinterWithContext(ctx).Printf(format, args...)
}

// PrintlnWithContext 使用上下文语言进行 Println 格式化
func (m *Manager) PrintlnWithContext(ctx context.Context, args ...any) (n int, err error) {
	return m.GetPrinterWithContext(ctx).Println(args...)
}

// FprintWithContext 使用上下文语言进行 Fprint 格式化
func (m *Manager) FprintWithContext(ctx context.Context, w io.Writer, args ...any) (n int, err error) {
	return m.GetPrinterWithContext(ctx).Fprint(w, args...)
}

// FprintfWithContext 使用上下文语言进行 Fprintf 格式化
func (m *Manager) FprintfWithContext(ctx context.Context, w io.Writer, format string, args ...any) (n int, err error) {
	return m.GetPrinterWithContext(ctx).Fprintf(w, format, args...)
}

// FprintlnWithContext 使用上下文语言进行 Fprintln 格式化
func (m *Manager) FprintlnWithContext(ctx context.Context, w io.Writer, args ...any) (n int, err error) {
	return m.GetPrinterWithContext(ctx).Fprintln(w, args...)
}
