// Package xtext 提供基于 golang.org/x/text 的国际化和本地化实现。
//
// 本包包含以下核心组件：
// - PrinterFactory: 翻译数据管理和 Printer 创建
// - Loader: gotext 格式翻译文件加载器
// - Source: 翻译源管理，支持单个文件或目录扫描
// - Printer: 基于 catalog 的消息格式化器
//
// 特性：
// - 支持标准 golang.org/x/text/catalog 机制
// - 支持 gotext 格式（.gotext.json 和 .gotext.jsonc）
// - 实现高效的并发访问和缓存
// - 支持目录扫描和文件监听
// - 提供灵活的翻译数据加载策略
//
// 基本用法：
//
//	// 创建工厂
//	factory := xtext.NewPrinterFactory(
//	    xtext.Fallback(msg.English),
//	    xtext.LogFunc(log.Printf),
//	)
//
//	// 加载翻译数据
//	factory.Reset("/path/to/translations")
//
//	// 创建打印机
//	printer, err := factory.CreatePrinter(msg.English)
package xtext

import (
	"bytes"
	"cmp"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"

	"go-slim.dev/infra/msg"
	"golang.org/x/sync/singleflight"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

// 确保 PrinterFactory 实现了 msg.PrinterFactory 接口
var _ msg.PrinterFactory = (*PrinterFactory)(nil)

// PrinterFactory 基于 golang.org/x/text 实现的打印机工厂。
//
// 这个工厂负责：
// 1. 管理翻译数据的加载和组织
// 2. 提供线程安全的 Printer 创建
// 3. 实现高效的缓存机制
// 4. 支持 gotext 格式翻译文件
//
// 线程安全：所有公共方法都是线程安全的。
//
// 核心组件：
// - builder: 全局 catalog.Builder，所有翻译数据都加载到这里
// - sources: 翻译源列表，按语言范围从小到大排序
// - printers: Printer 缓存，避免重复创建
// - sf: singleflight 组，防止并发创建相同的 Printer
type PrinterFactory struct {
	mu       sync.RWMutex               // 读写锁，保护并发访问
	fallback msg.Locale                 // 回退语言，当找不到匹配的语言时使用
	logFunc  msg.LogFunc                // 日志函数，用于记录调试和错误信息
	sources  []*Source                  // 翻译源列表，按语言范围从小到大排序
	locales  msg.LocaleSet              // 语言集合，用于快速查找和匹配
	loaders  *LoaderRegistry            // 加载器注册表，支持多种文件格式
	builder  *catalog.Builder           // 全局 catalog.Builder，所有翻译数据都加载到这里
	printers map[msg.Locale]msg.Printer // Printer 缓存，key 为完整的 Locale（可能包含扩展信息）
	sf       singleflight.Group         // singleflight 组，用于避免重复创建 Printer
}

// options 包含 PrinterFactory 的配置选项
type options struct {
	baseDir  string          // 语言包目录
	fallback msg.Locale      // 回退语言
	logFunc  msg.LogFunc     // 日志函数
	loaders  *LoaderRegistry // 加载器注册表
}

// Option 定义 PrinterFactory 的配置选项函数类型
type Option func(*options)

// LogFunc 设置日志函数选项。
//
// 用于配置工厂的日志记录行为。如果不设置，则不记录日志。
//
// 参数 f: 日志函数
// 返回: 可用于 NewPrinterFactory 的选项
//
// 示例：
//
//	factory := xtext.NewPrinterFactory(
//	    xtext.LogFunc(func(msg string) {
//	        log.Printf("[XText] %s", msg)
//	    }),
//	)
func LogFunc(f msg.LogFunc) Option {
	return func(o *options) {
		o.logFunc = f
	}
}

// Fallback 设置回退语言选项。
//
// 当请求的语言不被支持时，会自动降级到指定的回退语言。
// 常见的回退语言选择：英语、用户首选语言等。
//
// 参数 f: 回退语言
// 返回: 可用于 NewPrinterFactory 的选项
//
// 示例：
//
//	factory := xtext.NewPrinterFactory(
//	    xtext.Fallback(msg.English), // 设置英语为回退语言
//	)
func Fallback(f msg.Locale) Option {
	return func(o *options) {
		o.fallback = f
	}
}

// BaseDir 设置翻译文件的根目录选项。
//
// 用于指定包含翻译文件的目录路径，工厂初始化时会自动加载该目录下的所有翻译文件。
// 相当于在创建工厂后调用 Reset(dir) 方法。
//
// 参数 dir: 翻译文件所在的根目录路径
// 返回: 可用于 NewPrinterFactory 的选项
//
// 示例：
//
//	factory := xtext.NewPrinterFactory(
//	    xtext.BaseDir("./locales"), // 自动加载 ./locales 目录下的翻译文件
//	    xtext.Fallback(msg.English),
//	)
func BaseDir(dir string) Option {
	return func(o *options) {
		o.baseDir = dir
	}
}

// Loaders 设置自定义的加载器注册表选项。
//
// 用于指定自定义的翻译文件加载器注册表，如果不设置，将使用默认的注册表。
// 可以通过此选项完全控制翻译文件的加载行为。
//
// 参数 l: 自定义的加载器注册表实例
// 返回: 可用于 NewPrinterFactory 的选项
//
// 示例：
//
//	registry := xtext.NewLoaderRegistry()
//	// registry 已经默认注册了 JSONLoader
//
//	factory := xtext.NewPrinterFactory(
//	    xtext.Loaders(registry),
//	    xtext.Fallback(msg.English),
//	)
func Loaders(l *LoaderRegistry) Option {
	return func(o *options) {
		o.loaders = l
	}
}

// NewPrinterFactory 创建 xtext 打印机工厂
func NewPrinterFactory(opts ...Option) *PrinterFactory {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	fallback := cmp.Or(o.fallback, msg.English)
	builder := catalog.NewBuilder()

	f := &PrinterFactory{
		fallback: fallback,
		logFunc:  o.logFunc,
		loaders:  o.loaders,
		builder:  builder,
		printers: make(map[msg.Locale]msg.Printer),
	}
	if f.logFunc == nil {
		f.logFunc = func(string) {} // discard
	}
	if f.loaders == nil {
		f.loaders = NewLoaderRegistry()
	}

	if o.baseDir != "" {
		f.Reset(o.baseDir)
	}

	return f
}

// Reset 重置工厂，清空所有加载的翻译数据
// 这个方法可以用于：
// 1. 清理内存中的翻译缓存
// 2. 重新加载所有翻译文件
// 3. 测试环境中的状态重置
//
// 注意：调用 Reset 后，所有之前创建的 Printer 仍然有效，
// 但它们使用的是重置前的 catalog 数据。如果需要最新的数据，
// 请重新创建 Printer 实例。
func (f *PrinterFactory) Reset(baseDir string, callbacks ...func(*catalog.Builder)) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.sources = f.loadSources(baseDir)
	f.locales = make(msg.LocaleSet, len(f.sources))
	f.builder = catalog.NewBuilder()
	f.printers = make(map[msg.Locale]msg.Printer)

	for i, s := range f.sources {
		f.locales[i] = s.locale
	}

	for _, callback := range callbacks {
		callback(f.builder)
	}
}

func (f *PrinterFactory) loadSources(baseDir string) []*Source {
	// 扫描目录，规则是：
	// - 文件 /baseDir/locale.gotext.json，创建包含该文件的 Source
	// - 文件 /baseDir/locale.gotext.jsonc，创建包含该文件的 Source
	// - 目录 /baseDir/locale/，扫描其中所有 .gotext.json 和 .gotext.jsonc 文件，创建 Source
	var sources []*Source

	if baseDir == "" {
		return sources
	}

	// 读取基础目录
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to read base directory %s: %v\n", baseDir, err)
		return sources
	}

	// 遍历基础目录中的条目
	for _, entry := range entries {
		fullPath := baseDir + "/" + entry.Name()

		if entry.IsDir() {
			locale, ok := parseBaseLocale(entry.Name())
			if !ok {
				// 处理无效的 locale 名称，记录警告信息
				fmt.Fprintf(os.Stderr, "Warning: invalid locale directory name: %s, skipping\n", entry.Name())
				continue
			}

			// 如果是目录，扫描其中的所有翻译文件
			entries := ScanDirectoryForEntries(fullPath, f.loaders)
			if len(entries) > 0 {
				source := NewSource(locale, entries)
				sources = append(sources, source)
			}
		} else {
			// 如果是文件，检查是否被加载器支持
			if loader, ok := f.loaders.GetLoaderForFile(fullPath); ok {
				// 从文件名推断 locale
				// 支持格式：locale.gotext.json 或 locale.gotext.jsonc
				name := entry.Name()
				var localeName string
				if strings.HasSuffix(name, ".gotext.json") {
					localeName = strings.TrimSuffix(name, ".gotext.json")
				} else if strings.HasSuffix(name, ".gotext.jsonc") {
					localeName = strings.TrimSuffix(name, ".gotext.jsonc")
				}

				if localeName != "" {
					locale, ok := parseBaseLocale(localeName)
					if !ok {
						// 处理无效的 locale 文件名，记录警告信息
						fmt.Fprintf(os.Stderr, "Warning: invalid locale file name: %s, skipping\n", name)
						continue
					}
					entries := []Entry{{file: fullPath, loader: loader}}
					source := NewSource(locale, entries)
					sources = append(sources, source)
				}
			}
		}
	}

	// 进行排序，比如：zh-Hans-CN、zh-Hans、zh-CN
	slices.SortFunc(sources, func(a, b *Source) int {
		// 注意小的排前面
		if a.locale.Contains(b.locale) {
			return -1
		}
		if b.locale.Contains(a.locale) {
			return 1
		}
		// 这里按字符字典序
		return bytes.Compare([]byte(a.locale), []byte(b.locale))
	})

	return sources
}

// CreatePrinter 实现 msg.PrinterFactory 接口
func (f *PrinterFactory) CreatePrinter(locale msg.Locale) (msg.Printer, error) {
	// 使用 singleflight 避免重复创建 Printer
	ch, err, _ := f.sf.Do(string(locale), func() (any, error) {
		// 检查缓存
		f.mu.RLock()
		if printer, exists := f.printers[locale]; exists {
			f.mu.RUnlock()
			return printer, nil
		}
		f.mu.RUnlock()

		// 加载翻译数据并创建 Printer
		printer, err := f.loadCatalogAndCreatePrinter(locale)
		if err != nil {
			return nil, err
		}

		// 存储到缓存
		f.mu.Lock()
		f.printers[locale] = printer
		f.mu.Unlock()

		return printer, nil
	})

	if err != nil {
		return nil, err
	}

	return ch.(msg.Printer), nil
}

// loadCatalogAndCreatePrinter 加载翻译数据并创建 Printer
func (f *PrinterFactory) loadCatalogAndCreatePrinter(locale msg.Locale) (msg.Printer, error) {
	f.mu.RLock()
	i := slices.IndexFunc(f.sources, func(s *Source) bool {
		return s.locale.Contains(locale)
	})
	fallback := f.fallback
	f.mu.RUnlock()

	if i == -1 {
		if !locale.Equal(fallback) {
			return f.loadCatalogAndCreatePrinter(fallback)
		}
		// 没有支持的语言，使用基本的 Printer
		return NewPrinter(locale, message.Catalog(catalog.NewBuilder()))
	}

	// 需要重新获取锁来访问 sources 和 builder
	f.mu.RLock()
	defer f.mu.RUnlock()

	// 先加载数据到 builder
	src := f.sources[i]
	src.SetLogFunc(f.logFunc)
	src.Load(f.builder)

	// 然后创建 Printer
	return NewPrinter(locale, message.Catalog(f.builder))
}

// SupportsLocale 实现 msg.PrinterFactory 接口
func (f *PrinterFactory) SupportsLocale(locale msg.Locale) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// 如果明确支持该语言，返回 true
	if f.locales.Contains(locale) {
		return true
	}

	// 如果没有明确支持的语言，或者该语言不匹配，则检查是否等于回退语言
	// 这是基于一个假设：工厂总是"支持"其回退语言，即使没有加载具体的翻译数据
	return locale.Equal(f.fallback)
}

// SupportedLocales 实现 msg.PrinterFactory 接口
func (f *PrinterFactory) SupportedLocales() msg.LocaleSet {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// 如果 locales 不为空，返回明确支持的语言
	if len(f.locales) > 0 {
		// 创建副本以防止修改影响原始 LocaleSet
		result := make(msg.LocaleSet, len(f.locales))
		copy(result, f.locales)
		return result
	}

	// 如果没有明确支持的语言，返回包含回退语言的 LocaleSet
	return msg.LocaleSet{f.fallback}
}

func (f *PrinterFactory) SetFallbackLocale(locale msg.Locale) msg.Locale {
	f.mu.Lock()
	defer f.mu.Unlock()

	old := f.fallback

	f.fallback = locale

	return old
}

func (f *PrinterFactory) GetFallbackLocale() msg.Locale {
	return f.fallback
}

func (f *PrinterFactory) SetTranslation(locale msg.Locale, key string, translation string) error {
	tag, err := language.All.Parse(locale.String())
	if err != nil {
		return err
	}

	f.builder.SetString(tag, key, translation)
	return nil
}

func (f *PrinterFactory) SetMacro(locale msg.Locale, name string, catamsg ...catalog.Message) error {
	tag, err := language.All.Parse(locale.String())
	if err != nil {
		return err
	}

	f.builder.SetMacro(tag, name, catamsg...)
	return nil
}

// RegisterLoader 注册新的文件加载器
func (f *PrinterFactory) RegisterLoader(loader Loader) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.loaders.Register(loader)
}

// UnregisterLoader 注销文件加载器
func (f *PrinterFactory) UnregisterLoader(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.loaders.Unregister(name)
}

// GetLoader 获取指定名称的加载器
func (f *PrinterFactory) GetLoader(name string) (Loader, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.loaders.GetLoader(name)
}

// GetSupportedExtensions 获取所有支持的文件扩展名
func (f *PrinterFactory) GetSupportedExtensions() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.loaders.GetSupportedExtensions()
}

// IsFileSupported 检查是否支持指定文件
func (f *PrinterFactory) IsFileSupported(filename string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.loaders.IsSupported(filename)
}
