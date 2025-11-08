package msg

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
)

func TestGetDefaultManager(t *testing.T) {
	t.Run("First call creates manager", func(t *testing.T) {
		// For testing, we'll rely on the existing behavior
		// Note: We cannot safely reset sync.Once in tests without modifying the package
		manager := GetDefaultManager()
		if manager == nil {
			t.Error("GetDefaultManager returned nil")
		}

		// Check that it has the expected default locale
		if manager.GetLocale() != English {
			t.Errorf("Default manager locale = %q, want %q",
				string(manager.GetLocale()), string(English))
		}
	})

	t.Run("Subsequent calls return same instance", func(t *testing.T) {
		manager1 := GetDefaultManager()
		manager2 := GetDefaultManager()

		if manager1 != manager2 {
			t.Error("GetDefaultManager returned different instances")
		}
	})

	t.Run("Thread safety", func(t *testing.T) {
		done := make(chan bool, 10)
		managers := make([]*Manager, 10)

		// Call GetDefaultManager concurrently
		for i := range 10 {
			go func(index int) {
				defer func() { done <- true }()
				managers[index] = GetDefaultManager()
			}(i)
		}

		// Wait for all goroutines
		for range 10 {
			<-done
		}

		// Verify all got the same instance
		for i := 1; i < 10; i++ {
			if managers[i] != managers[0] {
				t.Error("Concurrent calls returned different instances")
				break
			}
		}
	})
}

func TestSetDefaultManager(t *testing.T) {
	// Save original state
	originalManager := defaultManager

	defer func() {
		// Restore original state
		SetDefaultManager(originalManager)
	}()

	t.Run("Set custom manager", func(t *testing.T) {
		customManager := NewManager(ManagerConfig{
			Locale:  Chinese,
			LogFunc: func(string) {}, // Disable logging
		})

		SetDefaultManager(customManager)

		// Verify the manager was set
		manager := GetDefaultManager()
		if manager != customManager {
			t.Error("SetDefaultManager did not set the custom manager")
		}

		if manager.GetLocale() != Chinese {
			t.Errorf("Custom manager locale = %q, want %q",
				string(manager.GetLocale()), string(Chinese))
		}
	})

	t.Run("Set nil manager", func(t *testing.T) {
		SetDefaultManager(nil)

		// Next call should create a new manager
		manager := GetDefaultManager()
		if manager == nil {
			t.Error("GetDefaultManager returned nil after setting nil")
		}
	})

	t.Run("Thread safety", func(t *testing.T) {
		customManager := NewManager(ManagerConfig{
			Locale:  Japanese,
			LogFunc: func(string) {},
		})

		done := make(chan bool, 10)

		// Set manager concurrently
		for range 10 {
			go func() {
				defer func() { done <- true }()
				SetDefaultManager(customManager)
			}()
		}

		// Wait for all goroutines
		for range 10 {
			<-done
		}

		// Verify manager was set
		manager := GetDefaultManager()
		if manager.GetLocale() != Japanese {
			t.Error("Concurrent SetDefaultManager failed")
		}
	})
}

func TestGlobalLocale(t *testing.T) {
	// Save original state
	originalLocale := GetDefaultManager().GetLocale()

	defer func() {
		SetDefaultManager(NewManager(ManagerConfig{
			Locale:  originalLocale,
			LogFunc: func(string) {},
		}))
	}()

	t.Run("Set and get locale", func(t *testing.T) {
		SetLocale(Chinese)

		if GetLocale() != Chinese {
			t.Errorf("GetLocale() = %q, want %q", string(GetLocale()), string(Chinese))
		}

		SetLocale(French)

		if GetLocale() != French {
			t.Errorf("GetLocale() = %q, want %q", string(GetLocale()), string(French))
		}
	})

	t.Run("Locale affects formatting", func(t *testing.T) {
		SetLocale(English)
		result1 := Sprintf("Hello %s", "World")

		SetLocale(Chinese)
		result2 := Sprintf("Hello %s", "World")

		// Results should be the same since we're using the simple printer
		if result1 != result2 {
			t.Errorf("Simple printer results differ: %q != %q", result1, result2)
		}
	})
}

func TestGlobalFactory(t *testing.T) {
	t.Run("Set and get factory", func(t *testing.T) {
		customFactory := NewPrinterFactory()
		SetPrinterFactory(customFactory)

		// We can't directly verify the factory is set
		// but we can test that the factory works
		printer := GetPrinter()
		if printer == nil {
			t.Errorf("GetPrinter() should not be nil")
		}
		if printer == nil {
			t.Error("GetPrinter() returned nil")
		}
	})
}

func TestGlobalFormattingFunctions(t *testing.T) {
	t.Run("Sprint", func(t *testing.T) {
		result := Sprint("Hello", 42, true)
		expected := "Hello42true"
		if result != expected {
			t.Errorf("Sprint() = %q, want %q", result, expected)
		}
	})

	t.Run("Sprintf", func(t *testing.T) {
		result := Sprintf("Hello %s, number %d", "World", 42)
		expected := "Hello World, number 42"
		if result != expected {
			t.Errorf("Sprintf() = %q, want %q", result, expected)
		}
	})

	t.Run("Sprintln", func(t *testing.T) {
		result := Sprintln("Hello", "World")
		expected := "Hello World\n"
		if result != expected {
			t.Errorf("Sprintln() = %q, want %q", result, expected)
		}
	})

	t.Run("Printf", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		defer func() {
			w.Close()
			os.Stdout = oldStdout
		}()

		Printf("Hello %s\n", "World")
		w.Close()

		// Read output
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		output := string(buf[:n])

		expected := "Hello World\n"
		if output != expected {
			t.Errorf("Printf() output = %q, want %q", output, expected)
		}
	})

	t.Run("Println", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		defer func() {
			w.Close()
			os.Stdout = oldStdout
		}()

		Println("Hello", "World")
		w.Close()

		// Read output
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		output := string(buf[:n])

		expected := "Hello World\n"
		if output != expected {
			t.Errorf("Println() output = %q, want %q", output, expected)
		}
	})
}

func TestGlobalContextFunctions(t *testing.T) {
	t.Run("GetPrinterWithContext", func(t *testing.T) {
		ctx := context.Background()
		printer := GetPrinterWithContext(ctx)

		if printer == nil {
			t.Error("GetPrinterWithContext() returned nil")
		}

		if !printer.Locale().Equal(English) { // Default locale
			t.Errorf("GetPrinterWithContext() locale = %q, want English",
				string(printer.Locale()))
		}
	})

	t.Run("GetPrinterWithContext with locale", func(t *testing.T) {
		ctx := WithLocaleContext(context.Background(), Chinese)
		printer := GetPrinterWithContext(ctx)

		if !printer.Locale().Equal(Chinese) {
			t.Errorf("GetPrinterWithContext() locale = %q, want Chinese",
				string(printer.Locale()))
		}
	})

	t.Run("SprintfWithContext", func(t *testing.T) {
		ctx := WithLocaleContext(context.Background(), French)
		result := SprintfWithContext(ctx, "Hello %s", "World")
		expected := "Hello World"

		if result != expected {
			t.Errorf("SprintfWithContext() = %q, want %q", result, expected)
		}
	})

	t.Run("PrintfWithContext", func(t *testing.T) {
		ctx := WithLocaleContext(context.Background(), German)

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		defer func() {
			w.Close()
			os.Stdout = oldStdout
		}()

		PrintfWithContext(ctx, "Hello %s\n", "World")
		w.Close()

		// Read output
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		output := string(buf[:n])

		expected := "Hello World\n"
		if output != expected {
			t.Errorf("PrintfWithContext() output = %q, want %q", output, expected)
		}
	})
}

func TestGlobalUtilityFunctions(t *testing.T) {
	t.Run("WithLocale", func(t *testing.T) {
		calls := make([]string, 0)
		var mu sync.Mutex

		SetLocale(English)

		// This should use the current global locale (English)
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

		// Verify global locale is unchanged
		if GetLocale() != English {
			t.Errorf("Global locale changed to %q, should remain English", string(GetLocale()))
		}
	})

	t.Run("WithContext", func(t *testing.T) {
		ctx := WithLocaleContext(context.Background(), Japanese)
		calls := make([]string, 0)
		var mu sync.Mutex

		WithContext(ctx, func(printer Printer) {
			mu.Lock()
			calls = append(calls, string(printer.Locale()))
			mu.Unlock()
		})

		// Verify the call
		if len(calls) != 1 {
			t.Errorf("WithContext calls count = %d, want 1", len(calls))
		} else if calls[0] != "ja" {
			t.Errorf("WithContext call = %q, want ja", calls[0])
		}
	})
}

func TestGlobalThreadSafety(t *testing.T) {
	t.Run("Concurrent formatting", func(t *testing.T) {
		done := make(chan bool, 10)

		// Use global formatting functions concurrently
		for i := range 10 {
			go func() {
				defer func() { done <- true }()

				// Mix different formatting functions
				switch i % 4 {
				case 0:
					_ = Sprint("test", i)
				case 1:
					_ = Sprintf("test %d", i)
				case 2:
					_ = Sprintln("test", i)
				case 3:
					_ = GetPrinter()
				}
			}()
		}

		// Wait for all goroutines
		for range 10 {
			<-done
		}
	})

	t.Run("Concurrent locale changes", func(t *testing.T) {
		done := make(chan bool, 10)

		// Change locale concurrently
		for i := range 10 {
			go func() {
				defer func() { done <- true }()

				locales := []Locale{English, Chinese, Spanish, French}
				locale := locales[i%len(locales)]

				SetLocale(locale)
				_ = Sprintf("test %d", i)
			}()
		}

		// Wait for all goroutines
		for range 10 {
			<-done
		}
	})
}

func TestGlobalManagerReset(t *testing.T) {
	// Save original state
	originalManager := defaultManager

	defer func() {
		// Restore original state
		SetDefaultManager(originalManager)
	}()

	t.Run("Manager reset after SetDefaultManager", func(t *testing.T) {
		// Set custom manager
		customManager := NewManager(ManagerConfig{
			Locale:  Chinese,
			LogFunc: func(string) {},
		})
		SetDefaultManager(customManager)

		// Note: We cannot safely reset sync.Once in tests
		// Instead, we test the SetDefaultManager functionality
		manager := GetDefaultManager()
		if manager != customManager {
			t.Error("Manager should be the custom manager we set")
		}

		// Test that we can set a different manager
		anotherManager := NewManager(ManagerConfig{
			Locale:  French,
			LogFunc: func(string) {},
		})
		SetDefaultManager(anotherManager)

		newManager := GetDefaultManager()
		if newManager != anotherManager {
			t.Error("Manager should be updated to the new manager")
		}

		if newManager.GetLocale() != French {
			t.Errorf("Updated manager locale = %q, want French", string(newManager.GetLocale()))
		}
	})
}

func TestGlobalErrorHandling(t *testing.T) {
	t.Run("Nil locale handling", func(t *testing.T) {
		// This should not panic
		SetLocale(Locale(""))

		// Should be able to format
		result := Sprintf("test %s", "value")
		if result != "test value" {
			t.Errorf("Sprintf with empty locale = %q, want 'test value'", result)
		}
	})

	t.Run("Invalid format string", func(t *testing.T) {
		// This should not panic, but will produce output like "%!BAD FORMAT"
		result := Sprintf("%!BAD FORMAT")
		// We don't check exact output as it may vary by Go version
		if result == "" {
			t.Error("Sprintf with bad format returned empty string")
		}
	})
}

func TestGlobalExample(t *testing.T) {
	// This is essentially the same as Example but as a test
	// to ensure it continues to work

	// Set global default language
	SetLocale(ChineseSimplified)
	currentLocale := GetLocale()
	if !currentLocale.Equal(ChineseSimplified) {
		t.Errorf("Global locale = %q, want zh-Hans-CN", string(currentLocale))
	}

	// Use package-level formatting function
	result := Sprintf("Hello %s, today is %s", "World", "Monday")
	expected := "Hello World, today is Monday"
	if result != expected {
		t.Errorf("Formatted result = %q, want %q", result, expected)
	}

	// Use WithLocale to temporarily switch language
	called := false
	WithLocale(English, func(printer Printer) {
		called = true
		if !printer.Locale().Equal(English) {
			t.Errorf("WithLocale printer locale = %q, want en", string(printer.Locale()))
		}

		formatted := printer.Sprintf("Value: %d", 42)
		if formatted != "Value: 42" {
			t.Errorf("WithLocale formatted = %q, want 'Value: 42'", formatted)
		}
	})

	if !called {
		t.Error("WithLocale function was not called")
	}

	// Verify global locale is unchanged
	currentLocale2 := GetLocale()
	if !currentLocale2.Equal(ChineseSimplified) {
		t.Errorf("Global locale changed to %q, should remain zh-Hans-CN", string(currentLocale2))
	}
}

// Example 展示包级别函数的使用
func Example() {
	// 设置全局默认语言
	SetLocale(ChineseSimplified)
	fmt.Println("Global locale:", GetLocale())

	// 使用包级别的格式化函数
	result := Sprintf("Hello %s, today is %s", "World", "Monday")
	fmt.Println("Formatted:", result)

	// 使用 WithLocale 临时切换语言
	WithLocale(English, func(printer Printer) {
		fmt.Println("English:", printer.Sprintf("Value: %d", 42))
	})

	// Output:
	// Global locale: zh-Hans
	// Formatted: Hello World, today is Monday
	// English: Value: 42
}

// ExampleSetDefaultManager 展示自定义全局 Manager 的使用
func ExampleSetDefaultManager() {
	// 首先强制初始化默认 Manager 以重置状态
	_ = GetDefaultManager()

	// 创建自定义 Manager
	customManager := NewManager(ManagerConfig{
		Locale: Japanese,
		LogFunc: func(message string) {
			// 日志函数，这里不输出以保持示例简洁
		},
	})

	// 设置为全局默认 Manager
	SetDefaultManager(customManager)

	// 现在所有全局函数都使用自定义 Manager
	fmt.Println("Current locale:", GetLocale())
	result := Sprintf("Hello %s", "World")
	fmt.Println("Result:", result)

	// Output:
	// Current locale: ja
	// Result: Hello World
}
