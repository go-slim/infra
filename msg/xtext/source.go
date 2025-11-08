package xtext

import (
	"fmt"
	"os"

	"go-slim.dev/infra/msg"
	"golang.org/x/text/message/catalog"
)

// Entry 表示一个翻译文件条目，包含文件路径和对应的加载器。
//
// 这个结构体封装了翻译文件的基本信息：
// - file: 翻译文件的完整路径
// - loader: 用于加载该文件的加载器实例
//
// 示例：
//
//	Entry{
//	    file: "/path/to/translations.json",
//	    loader: xtext.NewJSONLoader(),
//	}
type Entry struct {
	file   string // 翻译文件的完整路径
	loader Loader // 文件加载器实例
}

// Source 表示一个翻译源，负责加载和管理特定语言的翻译数据。
//
// Source 是翻译加载的基本单元，每个 Source 对应一种语言，
// 可以包含多个翻译文件（通过 entries 字段）。
//
// 设计特点：
// 1. 语言特定：每个 Source 只处理一种语言的翻译
// 2. 文件聚合：支持单个文件或多个文件（通过 entries）
// 3. 延迟加载：只在需要时才加载翻译数据
// 4. 内存优化：加载完成后清空 entries 释放内存
// 5. 简单设计：不处理并发控制，由上层 PrinterFactory 管理
//
// 并发控制说明：
// - Source 本身不实现并发控制机制
// - 并发安全由 PrinterFactory 的 singleflight 机制保证
// - 这种设计避免了双重锁定的复杂性
// - 简化了 Source 的实现，提高了性能
//
// 使用示例：
//
//	// 创建包含多个文件的翻译源
//	entries := []xtext.Entry{
//	    {file: "en.json", loader: jsonLoader},
//	    {file: "en-ext.json", loader: jsonLoader},
//	}
//	source := xtext.NewSource(msg.English, entries)
type Source struct {
	locale  msg.Locale // 语言标识符
	entries []Entry    // 翻译文件条目列表
	logFunc msg.LogFunc
}

// NewSource 创建新的翻译源实例。
//
// 创建的 Source 可以立即使用，翻译数据会在第一次调用 load 方法时延迟加载。
//
// 参数 locale: 语言标识符
// 参数 entries: 翻译文件条目列表
// 返回: 初始化的 Source 实例
//
// 示例：
//
//	// 创建单个文件的翻译源
//	entry := xtext.Entry{file: "en.json", loader: jsonLoader}
//	source := xtext.NewSource(msg.English, []xtext.Entry{entry})
//
//	// 创建多个文件的翻译源
//	entries := []xtext.Entry{
//	    {file: "common.json", loader: jsonLoader},
//	    {file: "ui.json", loader: jsonLoader},
//	}
//	source := xtext.NewSource(msg.English, entries)
func NewSource(locale msg.Locale, entries []Entry) *Source {
	return &Source{
		locale:  locale,
		entries: entries,
	}
}

// SetLogFunc 设置日志函数用于记录加载过程中的信息和错误。
//
// 此方法允许为 Source 设置自定义的日志处理函数，用于记录
// 文件加载过程中的调试信息、警告和错误。如果不设置日志函数，
// 加载错误将不会被记录。
//
// 参数 f: 日志函数，接收格式化的消息字符串
//
// 使用示例：
//
//	source := xtext.NewSource(msg.English, entries)
//
//	// 设置日志函数记录到标准输出
//	source.SetLogFunc(func(msg string) {
//	    fmt.Printf("[Source] %s\n", msg)
//	})
//
//	// 设置日志函数记录到日志文件
//	source.SetLogFunc(func(msg string) {
//	    log.Printf("[Translation] %s", msg)
//	})
//
//	// 取消日志记录
//	source.SetLogFunc(nil)
//
// 线程安全性：
// - 此方法可以在任何时候调用
// - 设置后立即生效，影响后续的加载操作
// - 建议在开始加载前设置，避免遗漏日志
func (s *Source) SetLogFunc(f msg.LogFunc) {
	s.logFunc = f
}

// Load 将翻译数据加载到指定的 catalog.Builder 中。
//
// 这是 Source 的核心方法，负责：
// 1. 加载所有配置的翻译文件
// 2. 将数据合并到全局 catalog 中
// 3. 清空 entries 释放内存
//
// 注意：此方法不实现并发控制，并发安全由上层调用者保证。
// PrinterFactory 使用 singleflight 确保同一个 Source 不会被并发加载。
//
// 错误处理：
// - 单个文件加载失败不会影响其他文件的加载
// - 错误会通过设置的日志函数记录（如果有的话）
// - 此方法不会返回错误，只进行内部处理
//
// 参数 b: 目标 catalog.Builder，用于存储翻译数据
//
// 调用约定：
// - 应该由 PrinterFactory 的 singleflight 机制保护
// - 不应该在多个 goroutine 中同时调用同一个 Source 的 Load 方法
// - 重复调用（在 entries 已清空后）会直接返回成功
func (s *Source) Load(b *catalog.Builder) {
	if len(s.entries) > 0 {
		// 加载文件到工厂的 builder 中
		s.loadFileToBuilder(b)
		// 加载完成后清空 entries，释放不再需要的内存
		s.entries = nil
	}
}

// loadFileToBuilder 从文件加载翻译数据并合并到全局 builder 中。
//
// 这是一个内部辅助方法，遍历 Source 中配置的所有 Entry，
// 使用对应的加载器将翻译数据加载到指定的 catalog.Builder 中。
//
// 错误处理策略：
// - 单个文件加载失败不会影响其他文件的加载
// - 错误会通过设置的日志函数记录（如果有的话）
// - 此方法不会返回错误，继续处理剩余文件
//
// 参数 b: 目标 catalog.Builder
//
// 注意：此方法是内部方法，不应该被外部调用
func (s *Source) loadFileToBuilder(b *catalog.Builder) {
	if len(s.entries) == 0 {
		return
	}

	// 遍历所有预配置的 entries
	for _, entry := range s.entries {
		if err := s.loadSingleFile(entry.file, entry.loader, b); err != nil {
			// 记录错误但继续处理其他文件
			if s.logFunc != nil {
				s.logFunc(fmt.Sprintf("Error loading translation file %s: %v", entry.file, err))
			}
			continue
		}
	}
}

// loadSingleFile 加载单个翻译文件并处理错误。
//
// 这是一个内部辅助方法，负责读取单个翻译文件的内容，
// 并使用对应的加载器将数据加载到 catalog.Builder 中。
//
// 处理流程：
// 1. 读取文件内容到内存
// 2. 使用对应的加载器解析文件数据
// 3. 将翻译数据加载到指定的 catalog.Builder 中
//
// 参数 filePath: 翻译文件的完整路径
// 参数 loader: 负责解析该文件格式的加载器实例
// 参数 b: 目标 catalog.Builder，用于存储翻译数据
//
// 返回: 加载过程中遇到的错误，成功则返回 nil
//
// 注意：此方法是内部方法，不应该被外部调用
func (s *Source) loadSingleFile(filePath string, loader Loader, b *catalog.Builder) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read translation file %s: %w", filePath, err)
	}

	return loader.LoadToBuilder(filePath, data, b, s.locale)
}
