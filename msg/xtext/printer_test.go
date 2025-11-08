package xtext

import (
	"testing"

	"go-slim.dev/infra/msg"
)

func TestNewPrinter(t *testing.T) {
	t.Run("Valid locale", func(t *testing.T) {
		printer, err := NewPrinter(msg.English)
		if err != nil {
			t.Errorf("NewPrinter() error = %v", err)
		}
		if printer == nil {
			t.Error("NewPrinter() returned nil")
		}

		if !printer.Locale().Equal(msg.English) {
			t.Errorf("Printer locale = %q, want %q",
				string(printer.Locale()), string(msg.English))
		}
	})

	t.Run("Chinese locale", func(t *testing.T) {
		printer, err := NewPrinter(msg.Chinese)
		if err != nil {
			t.Errorf("NewPrinter() error = %v", err)
		}
		if printer == nil {
			t.Error("NewPrinter() returned nil")
		}

		if !printer.Locale().Equal(msg.Chinese) {
			t.Errorf("Printer locale = %q, want %q",
				string(printer.Locale()), string(msg.Chinese))
		}
	})

	t.Run("Complex locale", func(t *testing.T) {
		locale := msg.Locale("zh-Hans-CN")
		printer, err := NewPrinter(locale)
		if err != nil {
			t.Errorf("NewPrinter() error = %v", err)
		}
		if printer == nil {
			t.Error("NewPrinter() returned nil")
		}

		if !printer.Locale().Equal(locale) {
			t.Errorf("Printer locale = %q, want %q",
				string(printer.Locale()), string(locale))
		}
	})
}

func TestPrinter_Sprint(t *testing.T) {
	printer, _ := NewPrinter(msg.English)

	t.Run("Single argument", func(t *testing.T) {
		result := printer.Sprint("Hello")
		expected := "Hello"
		if result != expected {
			t.Errorf("Printer.Sprint() = %q, want %q", result, expected)
		}
	})

	t.Run("Multiple arguments", func(t *testing.T) {
		result := printer.Sprint("Hello", " ", "World")
		expected := "Hello World"
		if result != expected {
			t.Errorf("Printer.Sprint() = %q, want %q", result, expected)
		}
	})

	t.Run("Mixed types", func(t *testing.T) {
		result := printer.Sprint("Number:", 42, true)
		expected := "Number:42true"
		if result != expected {
			t.Errorf("Printer.Sprint() = %q, want %q", result, expected)
		}
	})
}

func TestPrinter_Sprintf(t *testing.T) {
	printer, _ := NewPrinter(msg.English)

	t.Run("Simple format", func(t *testing.T) {
		result := printer.Sprintf("Hello %s", "World")
		expected := "Hello World"
		if result != expected {
			t.Errorf("Printer.Sprintf() = %q, want %q", result, expected)
		}
	})

	t.Run("Multiple arguments", func(t *testing.T) {
		result := printer.Sprintf("%s %d %t", "Test", 42, true)
		expected := "Test 42 true"
		if result != expected {
			t.Errorf("Printer.Sprintf() = %q, want %q", result, expected)
		}
	})

	t.Run("No arguments", func(t *testing.T) {
		result := printer.Sprintf("Hello World")
		expected := "Hello World"
		if result != expected {
			t.Errorf("Printer.Sprintf() = %q, want %q", result, expected)
		}
	})
}

func TestPrinter_Sprintln(t *testing.T) {
	printer, _ := NewPrinter(msg.English)

	t.Run("Single argument", func(t *testing.T) {
		result := printer.Sprintln("Hello")
		expected := "Hello\n"
		if result != expected {
			t.Errorf("Printer.Sprintln() = %q, want %q", result, expected)
		}
	})

	t.Run("Multiple arguments", func(t *testing.T) {
		result := printer.Sprintln("Hello", "World")
		expected := "Hello World\n"
		if result != expected {
			t.Errorf("Printer.Sprintln() = %q, want %q", result, expected)
		}
	})
}

func TestPrinter_Locale(t *testing.T) {
	t.Run("English locale", func(t *testing.T) {
		printer, _ := NewPrinter(msg.English)
		if !printer.Locale().Equal(msg.English) {
			t.Errorf("Printer.Locale() = %q, want %q",
				string(printer.Locale()), string(msg.English))
		}
	})

	t.Run("Chinese locale", func(t *testing.T) {
		printer, _ := NewPrinter(msg.Chinese)
		if !printer.Locale().Equal(msg.Chinese) {
			t.Errorf("Printer.Locale() = %q, want %q",
				string(printer.Locale()), string(msg.Chinese))
		}
	})

	t.Run("Complex locale", func(t *testing.T) {
		locale := msg.Locale("zh-Hans-CN")
		printer, _ := NewPrinter(locale)
		if !printer.Locale().Equal(locale) {
			t.Errorf("Printer.Locale() = %q, want %q",
				string(printer.Locale()), string(locale))
		}
	})
}

func TestPrinter_DifferentLocales(t *testing.T) {
	locales := []msg.Locale{
		msg.English,
		msg.Chinese,
		msg.Spanish,
		msg.French,
		msg.German,
		msg.Japanese,
		msg.Korean,
		msg.Russian,
	}

	for _, locale := range locales {
		t.Run(string(locale), func(t *testing.T) {
			printer, err := NewPrinter(locale)
			if err != nil {
				t.Errorf("NewPrinter(%q) error = %v", string(locale), err)
			}
			if printer == nil {
				t.Errorf("NewPrinter(%q) returned nil", string(locale))
			}
			if !printer.Locale().Equal(locale) {
				t.Errorf("Printer.Locale() = %q, want %q",
					string(printer.Locale()), string(locale))
			}
		})
	}
}

func TestPrinter_ErrorHandling(t *testing.T) {
	t.Run("Invalid locale", func(t *testing.T) {
		invalidLocale := msg.Locale("invalid-locale-xyz-123-bad")
		printer, err := NewPrinter(invalidLocale)

		// Should still return a printer but with error
		if err == nil {
			t.Error("NewPrinter() should return error for invalid locale")
		}
		if printer == nil {
			t.Error("NewPrinter() should still return printer for invalid locale")
		}
	})
}
