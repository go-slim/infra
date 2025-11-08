package msg

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewManager(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		manager := NewManager(ManagerConfig{})
		if manager == nil {
			t.Error("NewManager returned nil")
		}

		// Should have default locale (English)
		if manager.GetLocale() != English {
			t.Errorf("Default locale = %q, want %q",
				string(manager.GetLocale()), string(English))
		}
	})

	t.Run("Custom configuration", func(t *testing.T) {
		config := ManagerConfig{
			Locale:  Chinese,
			LogFunc: func(msg string) {}, // Disable logging
			Factory: NewPrinterFactory(),
		}

		manager := NewManager(config)

		if manager.GetLocale() != Chinese {
			t.Errorf("Custom locale = %q, want %q",
				string(manager.GetLocale()), string(Chinese))
		}
	})
}

func TestManagerSetGetLocale(t *testing.T) {
	manager := NewManager(ManagerConfig{
		LogFunc: func(string) {}, // Disable logging
	})

	t.Run("Set and get locale", func(t *testing.T) {
		locales := []Locale{English, Chinese, Spanish, French, Japanese}

		for _, locale := range locales {
			manager.SetLocale(locale)
			currentLocale := manager.GetLocale()

			if !currentLocale.Equal(locale) {
				t.Errorf("SetLocale(%q) -> GetLocale() = %q",
					string(locale), string(currentLocale))
			}
		}
	})
}

func TestManagerGetPrinter(t *testing.T) {
	manager := NewManager(ManagerConfig{
		LogFunc: func(string) {},
	})

	t.Run("Get printer with current locale", func(t *testing.T) {
		manager.SetLocale(Chinese)

		printer := manager.GetPrinter()
		if printer == nil {
			t.Error("GetPrinter() returned nil")
		}

		if !printer.Locale().Equal(Chinese) {
			t.Errorf("Printer locale = %q, want %q",
				string(printer.Locale()), string(Chinese))
		}
	})

	t.Run("Get printer with specific locale", func(t *testing.T) {
		manager.SetLocale(English)

		printer := manager.GetPrinter(French)
		if printer == nil {
			t.Error("GetPrinter(French) returned nil")
		}

		if !printer.Locale().Equal(French) {
			t.Errorf("Printer locale = %q, want %q",
				string(printer.Locale()), string(French))
		}

		// Verify manager's locale hasn't changed
		if manager.GetLocale() != English {
			t.Errorf("Manager locale changed to %q, should remain English",
				string(manager.GetLocale()))
		}
	})

	t.Run("Caching", func(t *testing.T) {
		manager.SetLocale(Spanish)

		printer1 := manager.GetPrinter()
		printer2 := manager.GetPrinter()

		// Should return the same printer (cached)
		if printer1 != printer2 {
			t.Error("GetPrinter should return cached printer for same locale")
		}
	})
}

func TestManagerFormattingMethods(t *testing.T) {
	manager := NewManager(ManagerConfig{
		LogFunc: func(string) {},
	})

	t.Run("Sprint", func(t *testing.T) {
		result := manager.Sprint("Hello", 42, true)
		expected := "Hello42true"
		if result != expected {
			t.Errorf("Sprint() = %q, want %q", result, expected)
		}
	})

	t.Run("Sprintf", func(t *testing.T) {
		result := manager.Sprintf("Hello %s, number %d", "World", 42)
		expected := "Hello World, number 42"
		if result != expected {
			t.Errorf("Sprintf() = %q, want %q", result, expected)
		}
	})

	t.Run("Sprintln", func(t *testing.T) {
		result := manager.Sprintln("Hello", "World")
		expected := "Hello World\n"
		if result != expected {
			t.Errorf("Sprintln() = %q, want %q", result, expected)
		}
	})
}

func TestManagerUtilityMethods(t *testing.T) {
	manager := NewManager(ManagerConfig{
		LogFunc: func(string) {},
	})

	t.Run("WithLocale", func(t *testing.T) {
		manager.SetLocale(English)

		calls := make([]string, 0)
		var mu sync.Mutex

		// This should use the specified locale
		WithLocale(Chinese, func(printer Printer) {
			mu.Lock()
			calls = append(calls, string(printer.Locale()))
			mu.Unlock()
		})

		WithLocale(French, func(printer Printer) {
			mu.Lock()
			calls = append(calls, string(printer.Locale()))
			mu.Unlock()
		})

		// Verify the calls
		expected := []string{"zh-Hans-CN", "fr"}
		if len(calls) != len(expected) {
			t.Errorf("WithLocale calls count = %d, want %d", len(calls), len(expected))
		} else {
			for i, call := range calls {
				if call != expected[i] {
					t.Errorf("WithLocale call %d = %q, want %q", i, call, expected[i])
				}
			}
		}

		// Verify manager's locale is unchanged
		if manager.GetLocale() != English {
			t.Errorf("Manager locale changed to %q, should remain English",
				string(manager.GetLocale()))
		}
	})
}

// ExampleManager 展示简化后的 Manager 使用方法
func ExampleManager() {
	// 创建管理器
	manager := NewManager(ManagerConfig{
		Locale:  English,
		LogFunc: func(message string) {}, // 禁用日志输出以保持示例简洁
	})

	// 获取当前语言
	fmt.Println("Current locale:", manager.GetLocale())

	// 切换语言
	manager.SetLocale(ChineseSimplified)
	fmt.Println("Switched to:", manager.GetLocale())

	// 使用当前语言进行格式化
	result := manager.Sprintf("Hello %s", "World")
	fmt.Println("Formatted:", result)

	// Output:
	// Current locale: en
	// Switched to: zh-Hans
	// Formatted: Hello World
}

// Simple tests from manager_simple_test.go

func TestManager_Simple(t *testing.T) {
	t.Run("Create manager", func(t *testing.T) {
		manager := NewManager(ManagerConfig{})
		if manager == nil {
			t.Error("NewManager returned nil")
		}
	})

	t.Run("Set and get locale", func(t *testing.T) {
		manager := NewManager(ManagerConfig{})
		manager.SetLocale(English)
		if manager.GetLocale() != English {
			t.Errorf("Expected English, got %s", manager.GetLocale())
		}
	})

	t.Run("Get printer", func(t *testing.T) {
		manager := NewManager(ManagerConfig{})
		printer := manager.GetPrinter()
		if printer == nil {
			t.Error("GetPrinter returned nil")
		}
	})

	t.Run("Formatting", func(t *testing.T) {
		manager := NewManager(ManagerConfig{})
		result := manager.Sprintf("Hello %s", "World")
		expected := "Hello World"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})
}
