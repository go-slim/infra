package xtext

import (
	"os"
	"path/filepath"
	"testing"

	"go-slim.dev/infra/msg"
)

func TestNewPrinterFactory(t *testing.T) {
	t.Run("Default options", func(t *testing.T) {
		factory := NewPrinterFactory()
		if factory == nil {
			t.Fatal("NewPrinterFactory returned nil")
		}

		// Test with fallback
		if factory.GetFallbackLocale().Equal(msg.Locale("")) {
			t.Error("Default fallback locale should not be empty")
		}
	})

	t.Run("With options", func(t *testing.T) {
		factory := NewPrinterFactory(
			Fallback(msg.Chinese),
			LogFunc(func(msg string) {
				// Test that log function is called
			}),
		)

		if factory == nil {
			t.Fatal("NewPrinterFactory with options returned nil")
		}

		if !factory.GetFallbackLocale().Equal(msg.Chinese) {
			t.Errorf("Fallback locale = %q, want %q",
				string(factory.GetFallbackLocale()), string(msg.Chinese))
		}
	})
}

func TestPrinterFactory_Reset(t *testing.T) {
	t.Run("Reset with empty directory", func(t *testing.T) {
		factory := NewPrinterFactory()

		// Should not panic with empty directory
		factory.Reset("")

		// Should still support English as fallback
		if !factory.SupportsLocale(msg.English) {
			t.Error("Factory should support English after reset")
		}
	})

	t.Run("Reset with valid directory", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create test files
		testFile := filepath.Join(tempDir, "en.json")
		testData := `{
  "language": "en",
  "messages": [
    {
      "id": "hello",
      "message": "hello",
      "translation": "Hello World"
    }
  ]
}`
		if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		factory := NewPrinterFactory()

		// Reset should not panic
		factory.Reset(tempDir)

		// Should support English
		if !factory.SupportsLocale(msg.English) {
			t.Error("Factory should support English after reset with files")
		}
	})
}

func TestPrinterFactory_CreatePrinter(t *testing.T) {
	factory := NewPrinterFactory()

	t.Run("Create basic printer", func(t *testing.T) {
		printer, err := factory.CreatePrinter(msg.English)
		if err != nil {
			t.Errorf("CreatePrinter() error = %v", err)
		}
		if printer == nil {
			t.Error("CreatePrinter() returned nil")
		}
	})

	t.Run("Create printer for different locales", func(t *testing.T) {
		locales := []msg.Locale{
			msg.English,
			msg.Chinese,
			msg.Spanish,
			msg.French,
		}

		for _, locale := range locales {
			printer, err := factory.CreatePrinter(locale)
			if err != nil {
				t.Errorf("CreatePrinter(%q) error = %v", string(locale), err)
			}
			if printer == nil {
				t.Errorf("CreatePrinter(%q) returned nil", string(locale))
			}
		}
	})

	t.Run("Create printer for unsupported locale", func(t *testing.T) {
		// Should still work with fallback
		printer, err := factory.CreatePrinter(msg.Locale("unsupported-locale"))
		if err != nil {
			t.Errorf("CreatePrinter() for unsupported locale error = %v", err)
		}
		if printer == nil {
			t.Error("CreatePrinter() for unsupported locale returned nil")
		}
	})
}

func TestPrinterFactory_SupportsLocale(t *testing.T) {
	factory := NewPrinterFactory()

	t.Run("Support common locales", func(t *testing.T) {
		locales := []msg.Locale{
			msg.English,
			msg.Chinese,
			msg.Spanish,
			msg.French,
			msg.German,
			msg.Japanese,
		}

		for _, locale := range locales {
			if !factory.SupportsLocale(locale) {
				t.Errorf("SupportsLocale(%q) returned false", string(locale))
			}
		}
	})

	t.Run("Support complex locales", func(t *testing.T) {
		complexLocales := []msg.Locale{
			msg.Locale("zh-Hans-CN"),
			msg.Locale("en-US"),
			msg.Locale("fr-FR"),
			msg.Locale("de-DE"),
		}

		for _, locale := range complexLocales {
			if !factory.SupportsLocale(locale) {
				t.Errorf("SupportsLocale(%q) returned false for complex locale", string(locale))
			}
		}
	})
}

func TestPrinterFactory_FallbackLocale(t *testing.T) {
	factory := NewPrinterFactory()

	t.Run("Set and get fallback", func(t *testing.T) {
		_ = factory.GetFallbackLocale()

		newFallback := msg.Chinese
		factory.SetFallbackLocale(newFallback)

		if !factory.GetFallbackLocale().Equal(newFallback) {
			t.Errorf("Fallback locale = %q, want %q",
				string(factory.GetFallbackLocale()), string(newFallback))
		}

		// Test return value
		old := factory.SetFallbackLocale(msg.Spanish)
		if !old.Equal(newFallback) {
			t.Errorf("Returned fallback = %q, want %q", string(old), string(newFallback))
		}
	})

	t.Run("Create printer with fallback", func(t *testing.T) {
		factory.SetFallbackLocale(msg.English)

		// Create printer for unsupported locale should use fallback
		printer, err := factory.CreatePrinter(msg.Locale("completely-unsupported-locale"))
		if err != nil {
			t.Errorf("CreatePrinter() with fallback error = %v", err)
		}
		if printer == nil {
			t.Error("CreatePrinter() with fallback returned nil")
		}
	})
}

func TestPrinterFactory_SupportedLocales(t *testing.T) {
	factory := NewPrinterFactory()

	t.Run("Get supported locales", func(t *testing.T) {
		supported := factory.SupportedLocales()

		// Should have some supported locales
		if len(supported) == 0 {
			t.Error("SupportedLocales() should not be empty")
		}

		// Should include common locales
		if !supported.Contains(msg.English) {
			t.Error("English should be in supported locales")
		}
	})

	t.Run("Locale set operations", func(t *testing.T) {
		supported := factory.SupportedLocales()

		// Test contains
		if !supported.Contains(msg.English) {
			t.Error("Contains(English) should return true")
		}

		// Test slice representation
		slice := supported.Slice()
		if len(slice) == 0 {
			t.Error("Slice() should not return empty slice")
		}
	})
}

func TestPrinterFactory_ConcurrentUsage(t *testing.T) {
	factory := NewPrinterFactory()

	t.Run("Concurrent printer creation", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()

				locale := msg.Locale("test-locale-" + string(rune('A'+id)))
				printer, err := factory.CreatePrinter(locale)
				if err != nil {
					t.Errorf("CreatePrinter() error in goroutine %d: %v", id, err)
					return
				}
				if printer == nil {
					t.Errorf("Printer is nil in goroutine %d", id)
					return
				}
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// Integration tests from xtext_test.go

func TestPrinterFactoryBasicFunctionality(t *testing.T) {
	// 创建临时目录和文件
	tempDir := t.TempDir()

	// 创建测试翻译文件
	testFile := tempDir + "/en.json"
	testData := `{
  "language": "en",
  "messages": [
    {
      "id": "hello",
      "message": "hello",
      "translation": "Hello World"
    },
    {
      "id": "goodbye",
      "message": "goodbye",
      "translation": "Goodbye"
    }
  ]
}`

	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 创建工厂
	factory := NewPrinterFactory()

	// 重置到指定目录
	factory.Reset(tempDir)

	// 测试创建 Printer
	printer, err := factory.CreatePrinter(msg.Locale("en"))
	if err != nil {
		t.Fatalf("Failed to create printer: %v", err)
	}
	if printer == nil {
		t.Fatal("Printer should not be nil")
	}

	// 测试基本功能
	result := printer.Sprintf("Hello %s", "World")
	if result != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", result)
	}
}

func TestInvalidLocaleHandling(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建无效的 locale 目录名
	invalidDir := tempDir + "/invalid@locale"
	if err := os.Mkdir(invalidDir, 0755); err != nil {
		t.Fatalf("Failed to create invalid locale directory: %v", err)
	}

	// 重定向 stderr 来捕获警告信息
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// 创建工厂并重置到包含无效 locale 的目录
	factory := NewPrinterFactory()
	factory.Reset(tempDir)

	// 恢复 stderr
	w.Close()
	os.Stderr = oldStderr

	// 读取警告信息
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	warningMsg := string(buf[:n])

	// 验证有警告信息关于无效的 locale 名称
	if len(warningMsg) == 0 {
		t.Error("Expected warning messages for invalid locale names")
	}

	// 验证警告信息包含预期的内容
	if !containsAny(warningMsg, []string{"invalid locale directory name", "invalid locale file name"}) {
		t.Errorf("Expected warning about invalid locale names, got: %s", warningMsg)
	}
}

func TestConcurrentPrinterCreation(t *testing.T) {
	// 创建临时目录和文件
	tempDir := t.TempDir()

	// 创建测试翻译文件
	testFile := tempDir + "/en.json"
	testData := `{
  "language": "en",
  "messages": [
    {
      "id": "hello",
      "message": "hello",
      "translation": "Hello"
    }
  ]
}`

	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 创建工厂
	factory := NewPrinterFactory()
	factory.Reset(tempDir)

	// 并发创建多个 Printer
	done := make(chan bool, 10)
	printers := make([]msg.Printer, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			printer, err := factory.CreatePrinter(msg.Locale("en"))
			if err != nil {
				t.Errorf("Failed to create printer %d: %v", index, err)
				done <- false
				return
			}
			printers[index] = printer
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	successCount := 0
	for i := 0; i < 10; i++ {
		if <-done {
			successCount++
		}
	}

	if successCount != 10 {
		t.Errorf("Expected all 10 printer creations to succeed, got %d", successCount)
	}

	// 验证所有创建的 Printer 都是同一个实例（由于 singleflight）
	for i := 1; i < 10; i++ {
		if printers[0] != printers[i] {
			t.Error("All printers should be the same instance due to singleflight")
			break
		}
	}
}

// Simple tests from simple_test.go

func TestPrinterFactory_Simple(t *testing.T) {
	t.Run("Create factory", func(t *testing.T) {
		factory := NewPrinterFactory()
		if factory == nil {
			t.Error("NewPrinterFactory returned nil")
		}
	})

	t.Run("Create printer", func(t *testing.T) {
		factory := NewPrinterFactory()

		printer, err := factory.CreatePrinter(msg.English)
		if err != nil {
			t.Errorf("CreatePrinter error: %v", err)
		}
		if printer == nil {
			t.Error("CreatePrinter returned nil")
		}
	})

	t.Run("Supports locale", func(t *testing.T) {
		factory := NewPrinterFactory()
		if !factory.SupportsLocale(msg.English) {
			t.Error("Factory should support English")
		}
	})

	t.Run("Set fallback locale", func(t *testing.T) {
		factory := NewPrinterFactory()
		old := factory.SetFallbackLocale(msg.Chinese)
		factory.SetFallbackLocale(old) // Restore
	})
}

// 辅助函数：检查字符串是否包含任意一个子串
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if len(substr) > 0 && len(s) >= len(substr) {
			// 简单的包含检查
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
