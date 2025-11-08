package msg

import (
	"sync"
	"sync/atomic"
)

// PrinterFactory 定义了翻译打印机的工厂接口。
//
// 此接口采用工厂模式，用于创建和管理 Printer 实例。通过工厂接口，
// 可以支持不同的翻译引擎实现（如 fmt、golang.org/x/text 等），
// 并提供统一的创建和管理接口。
//
// 主要功能：
// 1. 创建 Printer 实例：根据语言环境创建对应的打印机
// 2. 语言支持检查：验证是否支持指定的语言环境
// 3. 回退机制：当找不到匹配语言时提供回退选项
// 4. 缓存管理：支持 Printer 实例的缓存以提高性能
//
// 设计优势：
// - 解耦创建逻辑：将 Printer 的创建与使用分离
// - 支持多实现：可以轻松切换不同的翻译引擎
// - 统一接口：为不同实现提供一致的 API
// - 可扩展性：便于添加新的翻译引擎支持
type PrinterFactory interface {
	// CreatePrinter 为指定的语言环境创建 Printer 实例。
	//
	// 这是工厂接口的核心方法，负责根据语言环境创建
	// 对应的翻译打印机实例。工厂应该处理语言回退、
	// 缓存管理和错误处理等逻辑。
	//
	// 参数：
	//   locale: 目标语言环境，如 "zh-CN", "en-US" 等
	//
	// 返回：
	//   Printer: 配置了指定语言环境的打印机实例
	//   error: 创建过程中遇到的错误（如不支持的语言、资源加载失败等）
	//
	// 实现要求：
	// - 应该支持缓存机制以提高性能
	// - 需要处理语言回退逻辑
	// - 必须是线程安全的
	// - 应该验证语言环境的有效性
	//
	// 使用示例：
	//
	//	printer, err := factory.CreatePrinter(Chinese)
	//	if err != nil {
	//	    return fmt.Errorf("failed to create printer: %w", err)
	//	}
	//	message := printer.Sprintf("Hello, World!")
	CreatePrinter(locale Locale) (Printer, error)

	// SupportsLocale 检查工厂是否支持指定的语言环境。
	//
	// 此方法用于验证工厂对特定语言环境的支持程度。
	// 不同类型的工厂可能有不同的支持策略：
	//
	// 支持策略：
	// - 专业翻译引擎（如 xtext）：基于可用翻译文件进行精确匹配
	// - 基础格式化引擎（如 fmt）：支持所有语言的基本格式化
	// - 自定义实现：根据具体业务逻辑决定支持范围
	//
	// 参数：
	//   locale: 要检查的语言环境
	//
	// 返回：
	//   bool: 如果支持该语言环境则返回 true，否则返回 false
	//
	// 实现指导：
	// - 翻译引擎应基于可用翻译资源判断支持
	// - 格式化引擎通常支持所有语言
	// - 返回值应与 CreatePrinter 的行为保持一致
	//
	// 使用示例：
	//
	//	// 检查翻译支持
	//	if factory.SupportsLocale(French) {
	//	    printer, _ := factory.CreatePrinter(French)
	//	    message := printer.Sprintf("Welcome") // 可能有翻译
	//	} else {
	//	    // 使用回退语言
	//	    printer, _ = factory.CreatePrinter(English)
	//	    message := printer.Sprintf("Welcome") // 格式化但没有翻译
	//	}
	SupportsLocale(locale Locale) bool

	// SupportedLocales 返回工厂支持的语言环境集合。
	//
	// 此方法提供工厂的语言支持能力信息，用于界面展示、
	// 用于语言选择界面的展示、文档生成或调试信息。
	// 返回的列表应该按照优先级或字母顺序排列。
	//
	// 返回：
	//   []Locale: 支持的语言环境列表，可能为空
	//
	// 实现注意事项：
	// - 返回 nil 表示工厂没有明确的语言限制
	// - 空切片表示工厂明确声明不支持任何语言
	// - 列表应该去除重复项
	// - 建议按常用程度排序
	//
	// 使用示例：
	//
	//	supported := factory.SupportedLocales()
	//	fmt.Printf("Supported languages: %v\n", supported)
	//
	//	for _, locale := range supported {
	//	    if factory.SupportsLocale(locale) {
	//	        printer, _ := factory.CreatePrinter(locale)
	//	        // 使用打印机...
	//	    }
	//	}
	SupportedLocales() LocaleSet

	// SetFallbackLocale 设置回退语言环境。
	//
	// 当请求的语言环境不被支持时，工厂将使用此回退语言
	// 创建 Printer。这对于处理未知或未配置的语言很重要。
	// 方法会返回之前设置的回退语言。
	//
	// 参数：
	//   locale: 要设置的回退语言环境
	//
	// 返回：
	//   Locale: 之前设置的回退语言环境，如果之前没有设置则返回空字符串
	//
	// 使用场景：
	// - 应用程序启动时设置默认回退语言
	// - 根据用户偏好动态调整回退语言
	// - 在多租户环境中为不同租户设置不同回退
	//
	// 使用示例：
	//
	//	// 设置英语为回退语言
	//	old := factory.SetFallbackLocale(English)
	//	if old != "" {
	//	    fmt.Printf("Previous fallback was: %s\n", old)
	//	}
	//
	//	// 现在所有不支持的语言都会回退到英语
	//	printer, _ := factory.CreatePrinter(Locale("unsupported"))
	SetFallbackLocale(locale Locale) (old Locale)

	// GetFallbackLocale 获取当前设置的回退语言环境。
	//
	// 返回工厂当前使用的回退语言，用于了解当前的语言
	// 回退策略。如果没有设置回退语言，将返回空字符串。
	//
	// 返回：
	//   Locale: 当前的回退语言环境，如果未设置则返回空字符串
	//
	// 使用示例：
	//
	//	fallback := factory.GetFallbackLocale()
	//	if fallback == "" {
	//	    fmt.Println("No fallback locale configured")
	//	} else {
	//	    fmt.Printf("Current fallback: %s\n", fallback)
	//	}
	GetFallbackLocale() Locale
}

// simplePrinterFactory 基于 fmt 驱动的工厂实现。
//
// 这是一个包装器工厂，支持两种模式：
// 1. 自定义模式：使用外部提供的 PrinterFactory 实现
// 2. 内置模式：使用基于 Go fmt 包的简单实现
//
// 内置模式特点：
// - 支持所有语言的基本格式化功能（如数字、日期格式）
// - 不提供真正的翻译功能，仅支持消息格式化
// - SupportsLocale 对所有语言都返回 true，与 CreatePrinter 行为一致
// - 适合开发、测试或不需要翻译的场景
//
// 结构特点：
// - 线程安全：使用读写锁保护自定义工厂的切换
// - 缓存机制：内置模式下缓存 Printer 实例
// - 原子操作：回退语言使用原子值管理
// - 灵活设计：支持运行时切换底层实现
//
// 使用场景：
// - 作为默认工厂提供基本功能
// - 作为适配器包装其他翻译引擎
// - 在开发阶段使用，生产环境替换为专业引擎
// - 需要基本格式化但不需要翻译的应用
type simplePrinterFactory struct {
	mu       sync.RWMutex   // 读写锁，保护 custom 字段的并发访问
	custom   PrinterFactory // 可选的自定义工厂，为 nil 时使用内置实现
	fallback atomic.Value   // 原子值存储回退语言，类型为 Locale
	printers sync.Map       // 缓存映射，只在内置模式下使用，支持所有语言的 Printer
}

// NewPrinterFactory 创建新的 fmt 驱动工厂实例。
//
// 此函数创建一个工厂实例，可以选择性地包装一个自定义的
// PrinterFactory 实现。如果提供了自定义工厂，所有操作都将
// 委托给该工厂；否则将使用基于 fmt 包的内置实现。
//
// 参数：
//
//	custom: 可选的自定义 PrinterFactory，最多支持一个
//
// 返回：
//
//	PrinterFactory: 新创建的工厂实例
//
// 使用示例：
//
//	// 使用内置 fmt 实现
//	factory := NewPrinterFactory()
//
//	// 包装自定义工厂
//	customFactory := NewXTextPrinterFactory(config)
//	factory := NewPrinterFactory(customFactory)
//
//	// 现在可以通过统一接口使用不同实现
//	printer, err := factory.CreatePrinter(Chinese)
func NewPrinterFactory(custom ...PrinterFactory) PrinterFactory {
	f := &simplePrinterFactory{}
	if len(custom) > 0 {
		f.custom = custom[0]
	}
	return f
}

// SetCustom 设置自定义工厂。
//
// 此方法允许运行时切换底层工厂实现。切换时会保持当前的
// 回退语言设置，确保配置的一致性。这是一个线程安全的操作。
//
// 参数：
//
//	custom: 要设置的自定义工厂，可以为 nil 以切换回内置实现
//
// 线程安全性：
// - 使用写锁保护，与 loadCustom() 操作互斥
// - 原子操作确保切换的原子性
// - 不影响正在进行的其他操作
//
// 使用示例：
//
//	factory := NewPrinterFactory()
//
//	// 运行时切换到专业翻译引擎
//	proFactory := NewProfessionalTranslator()
//	factory.SetCustom(proFactory)
//
//	// 切换回内置实现
//	factory.SetCustom(nil)
func (f *simplePrinterFactory) SetCustom(custom PrinterFactory) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.custom = custom

	// 如果设置了自定义工厂，同步当前的回退语言设置
	if custom != nil {
		f.custom.SetFallbackLocale(f.GetFallbackLocale())
	}
}

// loadCustom 安全地加载当前的自定义工厂。
//
// 这是一个内部辅助方法，使用读锁确保在并发环境下
// 安全地访问 custom 字段。所有公共方法都应该通过
// 此方法访问 custom 字段，而不是直接访问。
//
// 返回：
//
//	PrinterFactory: 当前设置的自定义工厂，可能为 nil
//
// 线程安全性：
// - 使用读锁保护，与 SetCustom() 操作互斥
// - 确保获取到的值是完整的，不会被并发修改打断
//
// 注意：此方法返回后，custom 字段仍可能被其他 goroutine 修改，
// 但这对于使用模式是安全的，因为我们通常只读取其方法调用结果。
func (f *simplePrinterFactory) loadCustom() PrinterFactory {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.custom
}

// CreatePrinter 为指定的语言环境创建 Printer 实例。
//
// 此方法根据当前工作模式创建 Printer：
// 1. 自定义模式：委托给自定义工厂创建
// 2. 内置模式：创建基于 fmt 包的简单实现，并使用缓存
//
// 缓存策略：
// - 只在内置模式下使用缓存
// - 使用 sync.Map 实现线程安全的缓存
// - 缓存键为语言环境，值为 Printer 实例
// - 缓存永不失效，适合语言环境数量有限的应用
//
// 参数：
//
//	locale: 目标语言环境
//
// 返回：
//
//	Printer: 创建的打印机实例
//	error: 创建过程中的错误（内置实现不会出错）
//
// 性能特性：
// - 自定义模式：性能取决于底层实现
// - 内置模式：首次创建后缓存，后续调用 O(1) 复杂度
// - 内存使用：缓存会随语言环境数量线性增长
//
// 使用示例：
//
//	factory := NewPrinterFactory()
//
//	// 首次调用会创建并缓存
//	printer1, err := factory.CreatePrinter(Chinese)
//
//	// 后续调用会从缓存返回
//	printer2, err := factory.CreatePrinter(Chinese)
//	// printer1 == printer2 (相同的实例)
func (f *simplePrinterFactory) CreatePrinter(locale Locale) (Printer, error) {
	// 如果有自定义工厂，委托给自定义实现
	if c := f.loadCustom(); c != nil {
		return c.CreatePrinter(locale)
	}

	// 内置模式：检查缓存
	if cached, exists := f.printers.Load(locale); exists {
		return cached.(Printer), nil
	}

	// 缓存未命中，创建新的 Printer 实例
	printer := NewPrinter(locale)
	f.printers.Store(locale, printer)
	return printer, nil
}

// SupportsLocale 检查工厂是否支持指定的语言环境。
//
// 支持检查逻辑：
// 1. 自定义模式：委托给自定义工厂的 SupportsLocale 方法
// 2. 内置模式：支持所有语言的基本格式化功能
//
// 内置模式说明：
// - fmt 实现可以处理任何语言的格式化需求
// - 虽然不提供真正的翻译功能，但支持基本的消息格式化
// - 这与 CreatePrinter 的行为保持一致，确保 API 的一致性
//
// 参数：
//
//	locale: 要检查的语言环境
//
// 返回：
//
//	bool: 内置模式总是返回 true，自定义模式取决于实现
//
// 使用示例：
//
//	factory := NewPrinterFactory()
//
//	// 内置模式：支持所有语言
//	if factory.SupportsLocale(Chinese) {
//	    fmt.Println("Chinese formatting is supported")
//	}
//
//	if factory.SupportsLocale(Locale("zh-CN")) {
//	    fmt.Println("Simplified Chinese formatting is supported")
//	}
//
//	// 自定义模式：取决于具体实现
//	customFactory := NewXTextPrinterFactory(config)
//	factory = NewPrinterFactory(customFactory)
//	if factory.SupportsLocale(Chinese) {
//	    fmt.Println("Chinese translation is supported")
//	}
func (f *simplePrinterFactory) SupportsLocale(locale Locale) bool {
	// 自定义模式：委托给自定义实现
	if c := f.loadCustom(); c != nil {
		return c.SupportsLocale(locale)
	}

	// 内置模式：fmt 实现支持所有语言的基本格式化
	// 虽然不提供真正的翻译功能，但可以处理任何语言的格式化需求
	// 这与 CreatePrinter 的行为保持一致，后者也可以创建任何语言的 Printer
	return true
}

// SupportedLocales 返回工厂支持的所有语言环境列表。
//
// 返回策略：
// 1. 自定义模式：直接返回自定义工厂的支持列表
// 2. 内置模式：返回 nil，表示 fmt 实现不提供真正的本地化
//
// 内置模式说明：
// - fmt 包支持所有语言进行格式化，但不提供翻译功能
// - 返回 nil 表示没有语言限制，但不推荐用于生产环境
// - 生产环境应该使用专业的翻译引擎（如 xtext 包）
//
// 返回：
//
//	[]Locale: 支持的语言环境列表，nil 表示没有明确限制
//
// 使用建议：
// - 开发和测试：可以使用内置模式快速验证逻辑
// - 生产环境：强烈建议使用专业的翻译实现
//
// 使用示例：
//
//	factory := NewPrinterFactory()
//	supported := factory.SupportedLocales()
//
//	if supported == nil {
//	    fmt.Println("Using built-in fmt implementation (no real translation)")
//	} else {
//	    fmt.Printf("Supported locales: %v\n", supported)
//	}
func (f *simplePrinterFactory) SupportedLocales() LocaleSet {
	// 自定义模式：委托给自定义实现
	if c := f.loadCustom(); c != nil {
		return c.SupportedLocales()
	}

	// 内置模式：fmt 实现支持所有语言的基本格式化
	// 返回 nil LocaleSet 表示无限制支持
	return nil
}

// SetFallbackLocale 设置回退语言环境。
//
// 回退机制：
// - 当 CreatePrinter 被调用但传入的语言不被支持时使用
// - 自定义模式：回退逻辑由自定义工厂实现
// - 内置模式：创建时直接使用指定的语言（fmt 不做验证）
//
// 原子操作：
// - 使用 atomic.Value 确保线程安全
// - 返回之前设置的回退语言
// - 支持并发读写操作
//
// 参数：
//
//	locale: 要设置的回退语言环境
//
// 返回：
//
//	Locale: 之前设置的回退语言，如果之前未设置则返回空字符串
//
// 使用场景：
// - 应用启动时设置默认回退语言
// - 多语言应用中设置合理的回退策略
// - 动态调整回退语言以适应用户需求
//
// 使用示例：
//
//	factory := NewPrinterFactory()
//
//	// 设置英语为回退语言
//	old := factory.SetFallbackLocale(English)
//	fmt.Printf("Previous fallback: %s\n", old)
//
//	// 动态调整回退语言
//	factory.SetFallbackLocale(French)
func (f *simplePrinterFactory) SetFallbackLocale(locale Locale) (old Locale) {
	v := f.fallback.Swap(locale)
	if v == nil {
		return ""
	}
	return v.(Locale)
}

// GetFallbackLocale 获取当前设置的回退语言环境。
//
// 此方法返回工厂当前使用的回退语言，用于：
// - 了解当前的语言回退策略
// - 调试和日志记录
// - 配置验证和展示
//
// 返回值说明：
// - 空字符串：表示没有设置回退语言
// - 有效语言标签：如 "en", "zh-CN" 等
//
// 返回：
//
//	Locale: 当前的回退语言环境
//
// 使用示例：
//
//	factory := NewPrinterFactory()
//
//	// 检查初始状态
//	fallback := factory.GetFallbackLocale()
//	if fallback == "" {
//	    fmt.Println("No fallback locale set")
//	}
//
//	// 设置后检查
//	factory.SetFallbackLocale(English)
//	fallback = factory.GetFallbackLocale()
//	fmt.Printf("Current fallback: %s\n", fallback)
func (f *simplePrinterFactory) GetFallbackLocale() Locale {
	v := f.fallback.Load()
	if v == nil {
		return ""
	}
	return v.(Locale)
}
