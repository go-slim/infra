// Package msg 提供了国际化和本地化的核心功能，包括：
// - 语言标识符(Locale)的定义和操作
// - 消息格式化接口
// - 打印机(Printer)工厂和管理器
//
// 主要设计原则：
// 1. 支持标准的 BCP 47 语言标签格式
// 2. 提供层次化的语言匹配机制
// 3. 支持多种消息格式化器实现
// 4. 提供线程安全的并发访问
package msg

import (
	"strings"
	"sync"
)

// 常用的预定义 Locale
//
// 这些常量定义了最常用的语言标识符，遵循 BCP 47 标准。
// 语言标识符格式：<语言代码>-<脚本代码>-<地区代码>
// 例如：zh-Hans-CN (简体中文-中国大陆)
const (
	English            Locale = "en"      // 英语
	EnglishUS          Locale = "en-US"   // 美国英语
	EnglishGB          Locale = "en-GB"   // 英国英语
	Chinese            Locale = "zh"      // 中文
	ChineseSimplified  Locale = "zh-Hans" // 简体中文
	ChineseTraditional Locale = "zh-Hant" // 繁体中文
	ChineseCN          Locale = "zh-CN"   // 中国大陆中文
	ChineseTW          Locale = "zh-TW"   // 中国台湾中文
	ChineseHK          Locale = "zh-HK"   // 中国香港中文
	Japanese           Locale = "ja"      // 日语
	JapaneseJP         Locale = "ja-JP"   // 日本日语
	Korean             Locale = "ko"      // 韩语
	KoreanKR           Locale = "ko-KR"   // 韩国韩语
	French             Locale = "fr"      // 法语
	FrenchFR           Locale = "fr-FR"   // 法国法语
	German             Locale = "de"      // 德语
	GermanDE           Locale = "de-DE"   // 德国德语
	Spanish            Locale = "es"      // 西班牙语
	SpanishES          Locale = "es-ES"   // 西班牙西班牙语
	Russian            Locale = "ru"      // 俄语
	RussianRU          Locale = "ru-RU"   // 俄罗斯俄语
	Arabic             Locale = "ar"      // 阿拉伯语
	ArabicEG           Locale = "ar-EG"   // 埃及阿拉伯语
	Hindi              Locale = "hi"      // 印地语
	HindiIN            Locale = "hi-IN"   // 印度印地语
	Portuguese         Locale = "pt"      // 葡萄牙语
	Thai               Locale = "th"      // 泰语
	Vietnamese         Locale = "vi"      // 越南语
)

// Locale 表示一个语言标识符，基于 BCP 47 标准。
//
// BCP 47 语言标签格式：<语言代码>-<脚本代码>-<地区代码>-u-<扩展部分>-x-<私有部分>
//
// 示例：
//   - "en" (英语)
//   - "zh-Hans" (简体中文)
//   - "zh-Hans-CN" (简体中文-中国大陆)
//   - "en-US-u-ca-gregory" (美国英语，使用公历扩展)
//
// 注意：为了便于使用，支持将下划线格式 "en_US" 自动转换为标准格式 "en-US"。
type Locale string

// NewLocale 创建一个新的 Locale 实例，自动处理下划线格式转换。
//
// 参数 s 可以是以下格式之一：
//   - "en" (语言代码)
//   - "en_US" (下划线分隔，自动转换为 "en-US")
//   - "en-US" (标准格式)
//   - "zh_Hans_CN" (下划线分隔，自动转换为 "zh-Hans-CN")
//
// 返回标准格式的 Locale 实例。
func NewLocale(s string) Locale {
	return Locale(strings.ReplaceAll(s, "_", "-"))
}

// Base 返回 Locale 的基础语言标识符，去除扩展和私有部分。
//
// 返回的格式支持以下四种形式：
//   - "language" (语言代码，如 "en")
//   - "language-script" (语言-脚本，如 "zh-Hans")
//   - "language-region" (语言-地区，如 "en-US")
//   - "language-script-region" (语言-脚本-地区，如 "zh-Hans-CN")
//
// 返回值：
//   - (Locale, true): 成功解析出基础语言标识符
//   - ("", false): 解析失败，输入为空或格式无效
//
// 示例：
//   - "zh-Hans-CN-u-ca-gregory-x-private" → ("zh-Hans-CN", true)
//   - "en-US" → ("en-US", true)
//   - "" → ("", false)
func (l Locale) Base() (Locale, bool) {
	str := l.Language()
	if str == "" {
		return "", false
	}
	if script := l.Script(); script != "" {
		str += "-" + script
	}
	if region := l.Region(); region != "" {
		str += "-" + region
	}
	return Locale(str), true
}

// localeCache 用于缓存 Locale 解析结果，提高性能
var localeCache = sync.Map{} // map[string]localeParts

// localeParts 存储解析后的 Locale 各部分
type localeParts struct {
	slr, ext, pri string // slr:语言/脚本/地区, ext:扩展部分, pri:私有部分
}

// Parts 解析 Locale 字符串为三个主要部分。
//
// 返回值：
//   - slr: 语言/脚本/地区部分 (如 "zh-Hans-CN")
//   - ext: 扩展部分 (如 "ca-gregory-co-phonebk")
//   - pri: 私有部分 (如 "x-private")
//
// 解析规则：
//  1. 语言部分是必需的，使用连字符分隔
//  2. 扩展部分以 "u-" 开头，包含 Unicode 扩展
//  3. 私有部分以 "x-" 开头，包含私有标识
//
// 示例：
//   - "zh-Hans-CN-u-ca-gregory-x-private"
//     → ("zh-Hans-CN", "ca-gregory", "private")
//   - "en-US"
//     → ("en-US", "", "")
func (l Locale) Parts() (slr, ext, pri string) {
	str := string(l)
	if str == "" {
		return "", "", ""
	}

	// 尝试从缓存获取结果，提高性能
	if cached, ok := localeCache.Load(str); ok {
		parts := cached.(localeParts)
		return parts.slr, parts.ext, parts.pri
	}

	// 计算各部分
	slr, ext, pri = l.calculateParts()

	// 缓存结果以供后续使用
	localeCache.Store(str, localeParts{slr, ext, pri})

	return slr, ext, pri
}

// calculateParts 实际执行 Locale 字符串解析的逻辑。
//
// 这是一个内部方法，通过在字符串末尾添加连字符来简化边界检查。
// 解析算法：
//  1. 查找 "-u-" 标记扩展部分开始位置
//  2. 查找 "-x-" 标记私有部分开始位置
//  3. 根据位置信息分割字符串
func (l Locale) calculateParts() (slr, ext, pri string) {
	str := string(l)
	if str == "" {
		return "", "", ""
	}

	// 添加连字符简化边界检查，避免处理 "-u" 和 "-x" 的边界问题
	str = str + "-"

	u := strings.Index(str, "-u-")
	x := strings.Index(str, "-x-")

	if u == -1 && x == -1 {
		// 没有扩展或私有部分
		return str[:len(str)-1], "", ""
	}
	if u == -1 {
		// 只有私有部分，没有扩展部分
		return str[:x], "", str[x+3 : len(str)-1]
	}
	if x == -1 {
		// 只有扩展部分，没有私有部分
		return str[:u], str[u+3 : len(str)-1], ""
	}
	if u < x {
		// 扩展部分在私有部分之前
		return str[:u], str[u+3 : x], str[x+3 : len(str)-1]
	}
	// 私有部分在扩展部分之前（不太常见但合法）
	return str[:x], str[x+3 : u], str[u+3 : len(str)-1]
}

// Language 返回 Locale 的语言代码部分。
//
// 这是 Locale 的最基本组成部分，通常是两个字母的 ISO 639-1 代码。
//
// 示例：
//   - "en-US" → "en"
//   - "zh-Hans-CN" → "zh"
//   - "fr" → "fr"
func (l Locale) Language() string {
	str := string(l)
	i := strings.IndexByte(str, '-')
	if i == -1 {
		return str
	}
	return str[:i]
}

// Script 返回 Locale 的脚本代码部分。
//
// 脚本代码通常是四个字母的 ISO 15924 代码，用于指定文字系统。
//
// 示例：
//   - "zh-Hans-CN" → "Hans" (简体中文字符)
//   - "zh-Hant" → "Hant" (繁体中文字符)
//   - "en-US" → "" (英语不指定脚本)
func (l Locale) Script() string {
	base, _, _ := l.Parts()
	if i := strings.IndexByte(base, '-'); i != -1 {
		base = base[i+1:]
		if j := strings.IndexByte(base, '-'); j != -1 {
			return base[:j]
		}
		// 根据规定，脚本代码为4个字母，这里放宽条件检查
		if len(base) >= 4 {
			return base
		}
	}
	return ""
}

// Region 返回 Locale 的地区代码部分。
//
// 地区代码通常是两个字母的 ISO 3166-1 国家代码或三个字母的代码。
//
// 示例：
//   - "en-US" → "US" (美国)
//   - "zh-CN" → "CN" (中国)
//   - "en" → "" (英语不指定地区)
func (l Locale) Region() string {
	base, _, _ := l.Parts()
	if i := strings.IndexByte(base, '-'); i != -1 {
		base = base[i+1:]
		if j := strings.IndexByte(base, '-'); j != -1 {
			return base[j+1:]
		}
		// 地区代码通常为2-3个字符，脚本为4个字符
		if len(base) < 4 {
			return base
		}
	}
	return ""
}

// Extension 返回指定扩展键的值。
//
// 扩展键以 "u-" 开头，用于指定 Unicode 扩展。
// 常见的扩展键包括：
//   - "ca": 日历系统 (如 "gregory", "islamic")
//   - "co": 排序规则 (如 "phonebk", "pinyin")
//   - "nu": 数字系统 (如 "arab", "latn")
//   - "tz": 时区标识
//
// 参数 name: 扩展键名称
// 返回值: 对应的扩展值，如果不存在则返回空字符串
//
// 示例：
//   - "en-US-u-ca-gregory" 中，Extension("ca") → "gregory"
//   - "zh-CN-u-co-pinyin" 中，Extension("co") → "pinyin"
func (l Locale) Extension(name string) string {
	if name == "" {
		return ""
	}

	_, ext, _ := l.Parts()
	if ext == "" {
		return ""
	}

	// 解析扩展部分的键值对
	parts := strings.Split(ext, "-")
	for i := range parts {
		if parts[i] == name && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	return ""
}

// Extensions 返回所有扩展键值对的映射。
//
// 返回的映射包含所有以 "u-" 开头的扩展键及其对应的值。
// 如果没有扩展部分，则返回 nil。
//
// 示例：
//   - "en-US-u-ca-gregory-co-phonebk" → {"ca": "gregory", "co": "phonebk"}
//   - "zh-CN" → nil
func (l Locale) Extensions() map[string]string {
	_, ext, _ := l.Parts()
	if ext == "" {
		return nil
	}

	extMap := make(map[string]string)
	for pair := range strings.SplitSeq(ext, "-") {
		if i := strings.IndexByte(pair, '-'); i != -1 {
			extMap[pair[:i]] = pair[i+1:]
		}
	}
	return extMap
}

// PrivateUse 返回 Locale 的私有使用部分。
//
// 私有部分以 "x-" 开头，用于应用程序特定的标识。
// 这部分不在标准中定义，可以自由使用。
//
// 示例：
//   - "en-US-x-example" → "example"
//   - "zh-CN" → ""
func (l Locale) PrivateUse() string {
	_, _, pri := l.Parts()
	return pri
}

// Equal 检查两个 Locale 是否完全相等。
//
// 这是严格的字符串比较，包括语言、脚本、地区、扩展和私有部分。
//
// 示例：
//   - "en-US".Equal("en-US") → true
//   - "en-US".Equal("en") → false
//   - "en-US".Equal("en-GB") → false
func (l Locale) Equal(o Locale) bool {
	return l.String() == o.String()
}

// BaseEqual 检查两个 Locale 的基础部分是否相等。
//
// 基础部分包括语言、脚本和地区，不包括扩展和私有部分。
//
// 示例：
//   - "en-US-u-ca-gregory".BaseEqual("en-US") → true
//   - "en-US".BaseEqual("en") → false
//   - "zh-Hans-CN".BaseEqual("zh-Hans") → false
func (l Locale) BaseEqual(o Locale) bool {
	lb, _, _ := l.Parts()
	ob, _, _ := o.Parts()
	return lb == ob
}

// Contains 检查当前 Locale 是否包含另一个 Locale。
//
// 支持层次化匹配机制，遵循从通用到具体的层次：
//
//	语言 → 语言-脚本 → 语言-脚本-地区
//
// 包含规则：
//  1. 完全匹配：两个 Locale 完全相同
//  2. 更通用的包含更具体的：如 "zh" 包含 "zh-Hans-CN"
//  3. 更具体的不能包含更通用的：如 "zh-Hans-CN" 不包含 "zh"
//  4. 相同具体程度下，组件必须匹配：
//     - 相同具体程度时，地区必须相同：如 "zh-CN" 不包含 "zh-TW"
//     - 相同具体程度时，脚本必须相同：如 "zh-Hans" 不包含 "zh-Hant"
//  5. 不同语言不包含：如 "zh-CN" 不包含 "en-US"
//
// 示例：
//   - "zh" 包含 "zh", "zh-Hans", "zh-CN", "zh-Hans-CN" (更通用包含更具体)
//   - "zh-Hans" 包含 "zh-Hans", "zh-Hans-CN", "zh-Hans-SG" (相同脚本)
//   - "zh-CN" 包含 "zh-CN", "zh-Hans-CN", "zh-Latn-CN" (相同地区)
//   - "zh-Hans-CN" 不包含 "zh", "zh-Hans", "zh-CN" (更具体不包含更通用)
//   - "zh-CN" 不包含 "zh-TW", "zh-SG" (地区不同)
//   - "zh-Hans" 不包含 "zh-Hant" (脚本不同)
//
// 注意：这个方法实现了 BCP 47 中定义的 locale 匹配规则，
// 用于在查找合适的本地化资源时进行降级匹配。
func (l Locale) Contains(o Locale) bool {
	// 处理空字符串情况
	if l == "" && o == "" {
		return true
	}
	if l == "" || o == "" {
		return false
	}

	// 完全匹配
	if l.String() == o.String() {
		return true
	}

	// 语言不同，不能包含
	if l.Language() != o.Language() {
		return false
	}

	// 获取组成部分
	lScript, lRegion := l.Script(), l.Region()
	oScript, oRegion := o.Script(), o.Region()

	// 计算具体程度（组件数量）
	lSpecificity := 0
	if lScript != "" {
		lSpecificity++
	}
	if lRegion != "" {
		lSpecificity++
	}

	oSpecificity := 0
	if oScript != "" {
		oSpecificity++
	}
	if oRegion != "" {
		oSpecificity++
	}

	// 更具体的不能包含更通用的
	if lSpecificity > oSpecificity {
		return false
	}

	// 相同具体程度下，组件必须匹配
	if lSpecificity == oSpecificity {
		// 脚本必须相同
		if lScript != oScript {
			return false
		}
		// 地区必须相同
		if lRegion != oRegion {
			return false
		}
		// 相同具体程度且组件都相同，应该已经在前面的完全匹配中处理了
		return true
	}

	// l 更通用，o 更具体，检查组件兼容性
	// 如果 l 有脚本，o 也有脚本但不同，不能包含
	if lScript != "" && oScript != "" && lScript != oScript {
		return false
	}

	// 如果 l 有地区，o 也有地区但不同，不能包含
	if lRegion != "" && oRegion != "" && lRegion != oRegion {
		return false
	}

	// 通过所有检查，可以包含
	return true
}

// Compare 比较当前 Locale 与另一个 Locale，遵循标准的比较接口。
//
// 返回值：
//   - -1: 当前 Locale 在层次结构中排在另一个 Locale 之前（更具体）
//   - 0: 两个 Locale 完全相同
//   - 1: 当前 Locale 在层次结构中排在另一个 Locale 之后（更通用）
//
// 排序规则：
//  1. 首先按语言代码排序（字典序）
//  2. 相同语言下，按具体程度排序：具体 → 通用
//     - 语言-脚本-地区 < 语言-地区 < 语言-脚本 < 语言
//  3. 相同具体程度下，地区优先于脚本：
//     - 语言-地区 < 语言-脚本（因为地区固定而脚本可游离，且地区作为语言包名称出现几率更高）
//     - 其它情况按字典序
//
// 示例：
//   - "zh-Hans-CN".Compare("zh-Hans") → -1 (zh-Hans-CN 更具体：2个组件 < 1个组件)
//   - "zh-Hans".Compare("zh-Hans-CN") → 1 (zh-Hans 更通用：1个组件 > 2个组件)
//   - "zh-CN".Compare("zh-Hans") → -1 (zh-CN < zh-Hans：相同具体程度，地区优先)
//   - "en-US".Compare("en-Latn") → -1 (en-US < en-Latn：相同具体程度，地区优先)
//   - "zh-Hans".Compare("zh") → -1 (zh-Hans 更具体：1个组件 < 0个组件)
//   - "zh".Compare("en") → 1 (zh > en 字典序)
func (l Locale) Compare(o Locale) int {
	// 完全匹配
	if l.String() == o.String() {
		return 0
	}

	// 处理空字符串情况
	if l == "" && o != "" {
		return -1
	}
	if l != "" && o == "" {
		return 1
	}

	// 首先按语言代码比较
	lLang := l.Language()
	oLang := o.Language()
	if lLang != oLang {
		if lLang < oLang {
			return -1
		}
		return 1
	}

	// 语言相同，比较通用性
	lScript, lRegion := l.Script(), l.Region()
	oScript, oRegion := o.Script(), o.Region()

	// 计算具体程度
	lSpecificity := 0
	if lScript != "" {
		lSpecificity++
	}
	if lRegion != "" {
		lSpecificity++
	}

	oSpecificity := 0
	if oScript != "" {
		oSpecificity++
	}
	if oRegion != "" {
		oSpecificity++
	}

	// 具体程度优先：更具体的排在前面（-1）
	if lSpecificity != oSpecificity {
		if lSpecificity > oSpecificity {
			return -1
		}
		return 1
	}

	// 具体程度相同，地区优先于脚本
	// 因为地区是固定的，脚本可以游离在不同地区，且地区作为语言包名称出现几率更高
	if lRegion != "" && lScript == "" && oScript != "" && oRegion == "" {
		return -1 // 语言-地区 < 语言-脚本
	}
	if lScript != "" && lRegion == "" && oRegion != "" && oScript == "" {
		return 1 // 语言-脚本 > 语言-地区
	}

	// 相同具体程度下，按脚本比较
	if lScript != oScript {
		if lScript < oScript {
			return -1
		}
		return 1
	}

	// 脚本相同，按地区比较
	if lRegion != oRegion {
		if lRegion < oRegion {
			return -1
		}
		return 1
	}

	// 理论上不应该到达这里，因为前面已经处理了完全匹配
	return 0
}

// String 返回 Locale 的字符串表示。
//
// 这是实现 fmt.Stringer 接口的方法，允许 Locale 直接用于字符串格式化。
//
// 示例：
//   - fmt.Printf("Locale: %s", locale) // locale 为 Locale 类型
func (l Locale) String() string {
	return string(l)
}
