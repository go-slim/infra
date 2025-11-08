// Package xtext 提供了翻译文件加载器的实现。
//
// 本包支持 gotext 标准的 JSON 格式翻译文件：
// - .gotext.json: 标准的 JSON 格式翻译文件
// - .gotext.jsonc: 支持注释的 JSON 格式文件
//
// 设计特点：
// 1. 专注于 gotext 格式：严格支持 gotext 语言包格式
// 2. 扩展名明确：使用 .gotext.json 和 .gotext.jsonc 扩展名
// 3. 错误处理：完善的错误处理和日志记录
// 4. 并发安全：加载器是线程安全的
package xtext

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/tidwall/jsonc"
	"go-slim.dev/infra/msg"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
)

// Loader 定义翻译文件加载器接口。
//
// Loader 负责将 gotext 格式的翻译文件解析为 catalog.Builder 中的翻译数据。
type Loader interface {
	// Name 返回加载器的名称，用于标识和日志记录。
	Name() string

	// Extensions 返回支持的文件扩展名列表。
	// 扩展名格式为完整的后缀，如 ".gotext.json", ".gotext.jsonc"。
	Extensions() []string

	// CanLoad 检查是否可以加载指定文件。
	// 通过检查文件扩展名来判断加载器是否支持该文件。
	CanLoad(filename string) bool

	// LoadToBuilder 加载翻译文件并直接写入到指定的 builder 中。
	//
	// 参数:
	//   filename: 文件路径
	//   data: 文件内容字节数组
	//   builder: 目标 catalog.Builder，翻译数据将写入这里
	//   locale: 语言标识符，用于设置翻译的语言标签
	//
	// 返回: 加载过程中遇到的错误
	LoadToBuilder(filename string, data []byte, builder *catalog.Builder, locale msg.Locale) error
}

// JSONLoader 实现 gotext JSON 格式的加载器。
//
// 支持的文件格式：
// - .gotext.json: 标准 JSON 格式
// - .gotext.jsonc: 支持注释的 JSON 格式
//
// gotext 标准格式示例：
//
//	{
//	  "language": "zh-CN",
//	  "messages": [
//	    {
//	      "id": "Hello",
//	      "message": "Hello",
//	      "translation": "你好"
//	    }
//	  ]
//	}
type JSONLoader struct {
	name       string
	extensions []string
}

// NewJSONLoader 创建新的 JSON 加载器。
func NewJSONLoader() *JSONLoader {
	return &JSONLoader{
		name:       "JSON",
		extensions: []string{".gotext.json", ".gotext.jsonc"},
	}
}

// Name 返回加载器名称。
func (l *JSONLoader) Name() string {
	return l.name
}

// Extensions 返回支持的文件扩展名列表。
func (l *JSONLoader) Extensions() []string {
	return l.extensions
}

// CanLoad 检查是否可以加载指定文件。
func (l *JSONLoader) CanLoad(filename string) bool {
	lower := strings.ToLower(filename)
	for _, ext := range l.extensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// LoadToBuilder 加载 gotext JSON 格式的翻译文件并写入到指定的 builder 中。
func (l *JSONLoader) LoadToBuilder(filename string, data []byte, builder *catalog.Builder, locale msg.Locale) error {
	// 处理空文件
	if len(data) == 0 {
		return nil // 空文件是有效的，只是没有翻译内容
	}

	// 如果是 JSONC 格式，先转换为纯 JSON
	var jsonData []byte
	if strings.HasSuffix(strings.ToLower(filename), ".gotext.jsonc") {
		jsonData = jsonc.ToJSON(data)
	} else {
		jsonData = data
	}

	// 解析 JSON 数据
	var content map[string]any
	if err := json.Unmarshal(jsonData, &content); err != nil {
		return fmt.Errorf("failed to parse JSON translation file %s: %w", filename, err)
	}

	// 解析语言标签
	tag, err := language.Parse(string(locale))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to parse locale '%s', using default English: %v\n", locale, err)
		tag = language.English
	}

	// 处理 gotext 标准格式
	// 格式：{"language": "zh-CN", "messages": [{"id": "key", "message": "source", "translation": "target"}]}
	messages, ok := content["messages"].([]any)
	if !ok {
		return fmt.Errorf("invalid gotext format: missing 'messages' array in file %s", filename)
	}

	// 遍历 messages 数组
	for _, msg := range messages {
		msgMap, ok := msg.(map[string]any)
		if !ok {
			continue
		}

		id, hasID := msgMap["id"].(string)
		translation, hasTranslation := msgMap["translation"].(string)

		if hasID && hasTranslation && translation != "" {
			builder.SetString(tag, id, translation)
		}
	}

	return nil
}

// LoaderRegistry 加载器注册表，管理所有可用的加载器
type LoaderRegistry struct {
	loaders map[string]Loader // 按名称索引的加载器
	extMap  map[string]Loader // 按扩展名索引的加载器
	mu      sync.RWMutex      // 保护并发访问
	locked  bool              // 是否被锁定（不允许注册/注销）
}

// NewLoaderRegistry 创建新的加载器注册表
func NewLoaderRegistry() *LoaderRegistry {
	registry := &LoaderRegistry{
		loaders: make(map[string]Loader),
		extMap:  make(map[string]Loader),
	}

	// 注册默认加载器（JSONLoader 支持 .gotext.json 和 .gotext.jsonc）
	registry.Register(NewJSONLoader())

	return registry
}

// Register 注册新的加载器
func (r *LoaderRegistry) Register(loader Loader) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.locked {
		return fmt.Errorf("loader registry is locked, cannot register new loader")
	}

	r.loaders[loader.Name()] = loader

	// 建立扩展名映射
	for _, ext := range loader.Extensions() {
		r.extMap[strings.ToLower(ext)] = loader
	}

	return nil
}

// Unregister 注销加载器
func (r *LoaderRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.locked {
		return fmt.Errorf("loader registry is locked, cannot unregister loader")
	}

	if loader, exists := r.loaders[name]; exists {
		// 移除扩展名映射
		for _, ext := range loader.Extensions() {
			delete(r.extMap, strings.ToLower(ext))
		}
		// 移除加载器
		delete(r.loaders, name)
	}

	return nil
}

// GetLoader 获取指定名称的加载器
func (r *LoaderRegistry) GetLoader(name string) (Loader, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	loader, exists := r.loaders[name]
	return loader, exists
}

// GetLoaderForFile 获取支持指定文件的加载器
func (r *LoaderRegistry) GetLoaderForFile(filename string) (Loader, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 尝试匹配完整的扩展名（如 .gotext.json）
	lower := strings.ToLower(filename)
	for ext, loader := range r.extMap {
		if strings.HasSuffix(lower, ext) {
			return loader, true
		}
	}

	return nil, false
}

// GetAllLoaders 获取所有已注册的加载器
func (r *LoaderRegistry) GetAllLoaders() []Loader {
	r.mu.RLock()
	defer r.mu.RUnlock()
	loaders := make([]Loader, 0, len(r.loaders))
	for _, loader := range r.loaders {
		loaders = append(loaders, loader)
	}
	return loaders
}

// GetSupportedExtensions 获取所有支持的文件扩展名
func (r *LoaderRegistry) GetSupportedExtensions() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	extensions := make([]string, 0, len(r.extMap))
	for ext := range r.extMap {
		extensions = append(extensions, ext)
	}
	return extensions
}

// IsSupported 检查是否支持指定文件
func (r *LoaderRegistry) IsSupported(filename string) bool {
	_, exists := r.GetLoaderForFile(filename)
	return exists
}

// Lock 锁定注册表，不允许再注册或注销加载器
func (r *LoaderRegistry) Lock() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.locked = true
}

// Unlock 解锁注册表，允许注册或注销加载器
func (r *LoaderRegistry) Unlock() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.locked = false
}

// IsLocked 检查注册表是否被锁定
func (r *LoaderRegistry) IsLocked() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.locked
}
