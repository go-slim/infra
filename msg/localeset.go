package msg

import (
	"slices"
	"sort"
)

// LocaleSet 表示支持的语言环境集合，用于明确区分不同的语言支持策略。
//
// 此类型解决了传统 []Locale 无法表达的"无限制"语义，使 API 更加清晰和类型安全。
//
// 三种状态：
// - nil: 无限制支持，支持所有语言的基本格式化功能
// - []Locale{}: 空限制，明确声明不支持任何语言
// - [locales]: 有限制支持，只支持指定的语言列表
//
// 使用示例：
//
//	// 检查工厂的语言支持策略
//	supported := factory.SupportedLocales()
//	switch {
//	case supported.IsUnlimited():
//	    fmt.Println("支持所有语言的基本格式化")
//	case supported.IsEmpty():
//	    fmt.Println("不支持任何语言")
//	default:
//	    fmt.Printf("支持语言: %v", supported)
//	}
type LocaleSet []Locale

// IsUnlimited 检查是否为无限制的语言支持。
//
// 当 LocaleSet 为 nil 时，表示工厂支持所有语言的基本格式化功能，
// 但不提供真正的翻译。这通常用于基于 fmt 的基础实现。
//
// 返回：
//
//	bool: 如果是无限制支持则返回 true
func (ls LocaleSet) IsUnlimited() bool {
	return ls == nil
}

// IsEmpty 检查是否为空限制（明确声明不支持任何语言）。
//
// 当 LocaleSet 为非 nil 的空切片时，表示工厂明确声明不支持任何语言。
// 这与 IsUnlimited() 不同，后者表示支持所有语言。
//
// 返回：
//
//	bool: 如果是空限制则返回 true
func (ls LocaleSet) IsEmpty() bool {
	return !ls.IsUnlimited() && len(ls) == 0
}

// Sorted 返回排序后的 LocaleSet 副本。
//
// 此方法创建一个新的 LocaleSet，其中的 locales 按 BCP 47 标准排序：
// - 首先按语言代码排序
// - 相同语言下按具体程度排序（更具体的在前）
// - 相同具体程度下地区优先于脚本
//
// 返回：排序后的 LocaleSet 副本
func (ls LocaleSet) Sorted() LocaleSet {
	if ls.IsUnlimited() || ls.IsEmpty() || len(ls) == 1 {
		// 对于空、nil 或单元素的 LocaleSet，直接返回副本
		result := make(LocaleSet, len(ls))
		copy(result, ls)
		return result
	}

	// 检查是否已经排序
	isSorted := true
	for i := 1; i < len(ls); i++ {
		if ls[i-1].Compare(ls[i]) > 0 {
			isSorted = false
			break
		}
	}

	if isSorted {
		// 已经排序，直接返回副本
		result := make(LocaleSet, len(ls))
		copy(result, ls)
		return result
	}

	// 创建排序副本
	sorted := make(LocaleSet, len(ls))
	copy(sorted, ls)
	sort.Slice(sorted, func(i, j int) bool {
		// 安全比较，处理可能的无效 locale
		defer func() {
			if r := recover(); r != nil {
				// 如果 Compare panic，回退到字符串比较
			}
		}()

		result := sorted[i].Compare(sorted[j])
		return result < 0
	})

	return sorted
}

func (ls LocaleSet) Contains(locale Locale) bool {
	if ls.IsUnlimited() {
		return true
	}

	// 先排序以提高查找效率
	sorted := ls.Sorted()

	for _, supported := range sorted {
		// 双向层级包含检查：更通用的包含更具体的，更具体的也匹配更通用的
		if supported.Contains(locale) || locale.Contains(supported) {
			return true
		}
	}
	return false
}

// Contains 使用精确匹配检查是否包含指定的语言环境。
//
// 对于无限制的 LocaleSet，总是返回 true。
// 对于有限制的 LocaleSet，只检查是否存在完全相同的语言标签，
// 不支持父子语言匹配。
//
// 参数 locale: 要检查的语言环境
// 返回：如果精确包含该语言环境则返回 true
//
// 使用场景：
// - 配置验证：检查特定语言是否明确配置
// - 资源检查：验证特定语言资源是否存在
// - 权限控制：基于精确语言匹配的访问控制

// Slice 返回底层的语言切片的独立副本。
//
// 注意：如果 LocaleSet 是无限制的，此方法返回 nil。
// 在使用返回值前应该检查 IsUnlimited()。
// 返回的切片是独立的副本，修改不会影响原始 LocaleSet。
//
// 返回：底层语言切片的独立副本，无限制时返回 nil
func (ls LocaleSet) Slice() []Locale {
	return slices.Clone(ls)
}
