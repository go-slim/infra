package xtext

import (
	"fmt"
	"os"

	"go-slim.dev/infra/msg"
)

func parseBaseLocale(s string) (msg.Locale, bool) {
	// 检查基本的无效格式
	if s == "" {
		return "", false
	}

	// 检查是否包含无效字符 (@ 字符在 locale 名称中是无效的)
	for _, r := range s {
		if r == '@' {
			return "", false
		}
	}

	// 检查是否以破折号开头或结尾
	if len(s) > 0 && (s[0] == '-' || s[len(s)-1] == '-') {
		return "", false
	}

	// 检查是否有连续的破折号
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '-' && s[i+1] == '-' {
			return "", false
		}
	}

	// 如果通过了基本验证，使用 Base 方法
	locale, valid := msg.NewLocale(s).Base()

	// 再次检查结果是否为空
	if !valid || locale.String() == "" {
		return "", false
	}

	return locale, true
}

// ScanDirectoryForEntries 扫描目录中的所有 gotext 翻译文件，返回 Entry 列表。
// 忽略子目录，只处理 .gotext.json 和 .gotext.jsonc 文件。
func ScanDirectoryForEntries(dirPath string, loaders *LoaderRegistry) []Entry {
	var entries []Entry

	// 读取目录内容
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to read directory %s: %v\n", dirPath, err)
		return entries
	}

	// 如果 loaders 为 nil，直接返回空结果
	if loaders == nil {
		return entries
	}

	// 遍历目录中的文件
	for _, entry := range dirEntries {
		// 忽略子目录
		if entry.IsDir() {
			continue
		}

		// 构建文件完整路径
		filePath := dirPath + "/" + entry.Name()

		// 检查是否有支持的加载器
		loader, ok := loaders.GetLoaderForFile(filePath)
		if ok {
			entries = append(entries, Entry{
				file:   filePath,
				loader: loader,
			})
		}
	}

	return entries
}
