package xtext

import (
	"testing"

	"go-slim.dev/infra/msg"
	"golang.org/x/text/message/catalog"
)

func TestNewJSONLoader(t *testing.T) {
	loader := NewJSONLoader()

	if loader.Name() != "JSON" {
		t.Errorf("JSONLoader.Name() = %q, want 'JSON'", loader.Name())
	}

	exts := loader.Extensions()
	expected := []string{".gotext.json", ".gotext.jsonc"}
	if len(exts) != len(expected) {
		t.Errorf("JSONLoader.Extensions() length = %d, want %d", len(exts), len(expected))
	}
	for i, ext := range expected {
		if exts[i] != ext {
			t.Errorf("JSONLoader.Extensions()[%d] = %q, want %q", i, exts[i], ext)
		}
	}
}

func TestJSONLoaderCanLoad(t *testing.T) {
	loader := NewJSONLoader()

	tests := []struct {
		filename string
		expected bool
	}{
		{"test.gotext.json", true},
		{"test.gotext.jsonc", true},
		{"test.GOTEXT.JSON", true},  // Case insensitive
		{"test.GOTEXT.JSONC", true}, // Case insensitive
		{"test.json", false},        // Not .gotext.json
		{"test.jsonc", false},       // Not .gotext.jsonc
		{"test.txt", false},
		{"test", false},
		{"", false},
	}

	for _, tt := range tests {
		if loader.CanLoad(tt.filename) != tt.expected {
			t.Errorf("JSONLoader.CanLoad(%q) = %v, want %v",
				tt.filename, loader.CanLoad(tt.filename), tt.expected)
		}
	}
}

func TestJSONLoaderLoadToBuilder_StandardFormat(t *testing.T) {
	loader := NewJSONLoader()
	builder := catalog.NewBuilder()

	// gotext 标准格式
	data := []byte(`{
		"language": "zh-CN",
		"messages": [
			{
				"id": "Hello",
				"message": "Hello",
				"translation": "你好"
			},
			{
				"id": "Goodbye",
				"message": "Goodbye",
				"translation": "再见"
			}
		]
	}`)

	err := loader.LoadToBuilder("test.gotext.json", data, builder, msg.Locale("zh-CN"))
	if err != nil {
		t.Fatalf("LoadToBuilder() error = %v", err)
	}

	// TODO: Add assertions to verify translations were loaded correctly
	// This would require accessing the catalog internals or using the catalog
}

func TestJSONLoaderLoadToBuilder_InvalidFormat(t *testing.T) {
	loader := NewJSONLoader()
	builder := catalog.NewBuilder()

	// 缺少 messages 数组的无效格式
	data := []byte(`{
		"Hello": "你好",
		"Goodbye": "再见"
	}`)

	err := loader.LoadToBuilder("test.gotext.json", data, builder, msg.Locale("zh-CN"))
	if err == nil {
		t.Fatal("LoadToBuilder() should return error for format without 'messages' array")
	}
}

func TestJSONLoaderLoadToBuilder_JSONC(t *testing.T) {
	loader := NewJSONLoader()
	builder := catalog.NewBuilder()

	// JSONC 格式（带注释）
	data := []byte(`{
		// 语言标识
		"language": "zh-CN",
		// 翻译消息列表
		"messages": [
			{
				"id": "Hello",
				"message": "Hello",
				"translation": "你好" // 中文翻译
			}
		]
	}`)

	err := loader.LoadToBuilder("test.gotext.jsonc", data, builder, msg.Locale("zh-CN"))
	if err != nil {
		t.Fatalf("LoadToBuilder() error = %v", err)
	}

	// TODO: Add assertions to verify translations were loaded correctly
}

func TestJSONLoaderLoadToBuilder_EmptyFile(t *testing.T) {
	loader := NewJSONLoader()
	builder := catalog.NewBuilder()

	// 空文件应该被接受
	data := []byte(``)

	err := loader.LoadToBuilder("test.gotext.json", data, builder, msg.Locale("zh-CN"))
	if err != nil {
		t.Errorf("LoadToBuilder() with empty file should not error, got = %v", err)
	}
}

func TestJSONLoaderLoadToBuilder_InvalidJSON(t *testing.T) {
	loader := NewJSONLoader()
	builder := catalog.NewBuilder()

	// 无效的 JSON
	data := []byte(`{invalid json}`)

	err := loader.LoadToBuilder("test.gotext.json", data, builder, msg.Locale("zh-CN"))
	if err == nil {
		t.Error("LoadToBuilder() with invalid JSON should return error")
	}
}

func TestJSONLoaderLoadToBuilder_EmptyTranslation(t *testing.T) {
	loader := NewJSONLoader()
	builder := catalog.NewBuilder()

	// 包含空翻译的消息应该被跳过
	data := []byte(`{
		"messages": [
			{
				"id": "Hello",
				"message": "Hello",
				"translation": ""
			},
			{
				"id": "Goodbye",
				"message": "Goodbye",
				"translation": "再见"
			}
		]
	}`)

	err := loader.LoadToBuilder("test.gotext.json", data, builder, msg.Locale("zh-CN"))
	if err != nil {
		t.Fatalf("LoadToBuilder() error = %v", err)
	}

	// TODO: Verify that only "Goodbye" was loaded
}
