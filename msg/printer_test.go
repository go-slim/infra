package msg

import (
	"fmt"
	"testing"
)

func TestNewPrinter(t *testing.T) {
	t.Run("Create printer with locale", func(t *testing.T) {
		printer := NewPrinter(English)
		if printer == nil {
			t.Error("NewPrinter returned nil")
		}

		localizer, ok := printer.(Localizer)
		if !ok {
			t.Error("Printer does not implement Localizer interface")
		}

		if localizer.Locale() != English {
			t.Errorf("Printer locale = %q, want %q", localizer.Locale(), English)
		}
	})

	t.Run("Create printer with different locales", func(t *testing.T) {
		locales := []Locale{English, Chinese, Spanish, French, Japanese}

		for _, locale := range locales {
			printer := NewPrinter(locale)
			localizer := printer.(Localizer)
			if localizer.Locale() != locale {
				t.Errorf("Printer locale = %q, want %q", localizer.Locale(), locale)
			}
		}
	})
}

func TestPrinter_InterfaceCompliance(t *testing.T) {
	printer := NewPrinter(English)

	// Test that printer implements all required interfaces
	var _ Localizer = printer
	var _ Formatter = printer
	var _ WriterFormatter = printer
	var _ ConsoleFormatter = printer
	var _ Printer = printer

	t.Log("Printer implements all required interfaces")
}

func TestPrinter_Formatter(t *testing.T) {
	printer := NewPrinter(English)

	t.Run("Sprintf", func(t *testing.T) {
		result := printer.Sprintf("Hello %s", "World")
		expected := "Hello World"
		if result != expected {
			t.Errorf("Sprintf result = %q, want %q", result, expected)
		}
	})

	t.Run("Sprint", func(t *testing.T) {
		result := printer.Sprint("Hello", "World")
		expected := "HelloWorld"
		if result != expected {
			t.Errorf("Sprint result = %q, want %q", result, expected)
		}
	})

	t.Run("Sprintln", func(t *testing.T) {
		result := printer.Sprintln("Hello", "World")
		expected := "Hello World\n"
		if result != expected {
			t.Errorf("Sprintln result = %q, want %q", result, expected)
		}
	})
}

func TestPrinter_WriterFormatter(t *testing.T) {
	printer := NewPrinter(English)

	t.Run("Fprintf", func(t *testing.T) {
		var buf []byte
		writer := &testWriter{data: &buf}
		n, err := printer.Fprintf(writer, "Hello %s", "World")

		if err != nil {
			t.Errorf("Fprintf error = %v", err)
		}

		if n != 11 { // "Hello World" length
			t.Errorf("Fprintf returned %d bytes, want 11", n)
		}

		if string(buf) != "Hello World" {
			t.Errorf("Fprintf wrote %q, want \"Hello World\"", string(buf))
		}
	})

	t.Run("Fprint", func(t *testing.T) {
		var buf []byte
		writer := &testWriter{data: &buf}
		n, err := printer.Fprint(writer, "Hello", "World")

		if err != nil {
			t.Errorf("Fprint error = %v", err)
		}

		if n != 10 { // "HelloWorld" length
			t.Errorf("Fprint returned %d bytes, want 10", n)
		}

		if string(buf) != "HelloWorld" {
			t.Errorf("Fprint wrote %q, want \"HelloWorld\"", string(buf))
		}
	})

	t.Run("Fprintln", func(t *testing.T) {
		var buf []byte
		writer := &testWriter{data: &buf}
		n, err := printer.Fprintln(writer, "Hello", "World")

		if err != nil {
			t.Errorf("Fprintln error = %v", err)
		}

		if n != 12 { // "Hello World\n" length = 12 bytes
			t.Errorf("Fprintln returned %d bytes, want 12", n)
		}

		if string(buf) != "Hello World\n" {
			t.Errorf("Fprintln wrote %q, want \"Hello World\n\"", string(buf))
		}
	})
}

func TestPrinter_ConsoleFormatter(t *testing.T) {
	printer := NewPrinter(English)

	t.Run("Printf", func(t *testing.T) {
		// Capture stdout
		originalStdout := fmt.Sprintf // This is just a placeholder
		// In real testing, you would capture stdout using os.Stdout redirection
		_ = originalStdout

		// Since we can't easily capture stdout in this test, we'll just verify the method doesn't panic
		printer.Printf("Hello %s", "World")
	})

	t.Run("Print", func(t *testing.T) {
		printer.Print("Hello", "World")
	})

	t.Run("Println", func(t *testing.T) {
		printer.Println("Hello", "World")
	})
}

func TestPrinter_Localizer(t *testing.T) {
	printer := NewPrinter(ChineseSimplified)
	localizer := printer.(Localizer)

	if localizer.Locale() != ChineseSimplified {
		t.Errorf("Localizer locale = %q, want %q", localizer.Locale(), ChineseSimplified)
	}

	// Test that changing locale affects the localizer
	newPrinter := NewPrinter(English)
	newLocalizer := newPrinter.(Localizer)

	if newLocalizer.Locale() != English {
		t.Errorf("New localizer locale = %q, want %q", newLocalizer.Locale(), English)
	}
}

// testWriter is a simple writer implementation for testing
type testWriter struct {
	data *[]byte
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	*w.data = append(*w.data, p...)
	return len(p), nil
}

// ExampleNewPrinter 展示接口分离的优势
func ExampleNewPrinter() {
	printer := NewPrinter(English)

	// 可以只使用需要的功能
	var formatter Formatter = printer
	result := formatter.Sprintf("Value: %d", 42)
	fmt.Println("Formatted:", result)

	var localizer Localizer = printer
	fmt.Println("Locale:", localizer.Locale())

	// Output:
	// Formatted: Value: 42
	// Locale: en
}
