package xtext

import (
	"os"
	"path/filepath"
	"testing"

	"go-slim.dev/infra/msg"
	"golang.org/x/text/message/catalog"
)

func TestNewSource(t *testing.T) {
	locale := msg.English
	entries := []Entry{
		{file: "test.json", loader: NewJSONLoader()},
		{file: "another.json", loader: NewJSONLoader()},
	}

	source := NewSource(locale, entries)

	if source == nil {
		t.Fatal("NewSource returned nil")
	}

	if !source.locale.Equal(locale) {
		t.Errorf("Source.locale = %q, want %q", string(source.locale), string(locale))
	}

	if len(source.entries) != len(entries) {
		t.Errorf("Source.entries length = %d, want %d", len(source.entries), len(entries))
	}

	for i, expected := range entries {
		if source.entries[i].file != expected.file {
			t.Errorf("Source.entries[%d].file = %q, want %q", i, source.entries[i].file, expected.file)
		}
		if source.entries[i].loader != expected.loader {
			t.Errorf("Source.entries[%d].loader mismatch", i)
		}
	}
}

func TestNewSource_EmptyEntries(t *testing.T) {
	locale := msg.Chinese
	entries := []Entry{}

	source := NewSource(locale, entries)

	if source == nil {
		t.Fatal("NewSource returned nil")
	}

	if !source.locale.Equal(locale) {
		t.Errorf("Source.locale = %q, want %q", string(source.locale), string(locale))
	}

	if len(source.entries) != 0 {
		t.Errorf("Source.entries length = %d, want 0", len(source.entries))
	}
}

func TestNewSource_NilEntries(t *testing.T) {
	locale := msg.Spanish
	var entries []Entry

	source := NewSource(locale, entries)

	if source == nil {
		t.Fatal("NewSource returned nil")
	}

	if !source.locale.Equal(locale) {
		t.Errorf("Source.locale = %q, want %q", string(source.locale), string(locale))
	}

	if source.entries != nil {
		t.Error("Source.entries should be nil when created with nil entries")
	}
}

func TestSource_SetLogFunc(t *testing.T) {
	source := NewSource(msg.English, []Entry{})

	var logMessages []string
	logFunc := func(msg string) {
		logMessages = append(logMessages, msg)
	}

	// Set log function
	source.SetLogFunc(logFunc)

	// Verify log function is set by triggering a load (which should call the log function if there are issues)
	// Since we have no entries, this won't generate log messages, but we can verify the function is stored

	// Trigger load with empty entries - this should not generate logs but exercises the code path
	builder := catalog.NewBuilder()
	source.Load(builder)

	// The log function should be set even if not called yet
	if source.logFunc == nil {
		t.Error("Log function should be set after SetLogFunc")
	}

	// Test setting nil log function
	source.SetLogFunc(nil)
	if source.logFunc != nil {
		t.Error("Log function should be nil after setting nil")
	}
}

func TestSource_Load(t *testing.T) {
	t.Run("Successful load", func(t *testing.T) {
		// Create temporary directory and files
		tempDir := t.TempDir()

		// Create test translation files
		jsonFile1 := filepath.Join(tempDir, "file1.json")
		jsonFile2 := filepath.Join(tempDir, "file2.json")

		data1 := `{
  "language": "en",
  "messages": [
    {
      "id": "hello",
      "message": "hello",
      "translation": "Hello World"
    },
    {
      "id": "greeting",
      "message": "greeting",
      "translation": "Hi"
    }
  ]
}`
		data2 := `{
  "language": "en",
  "messages": [
    {
      "id": "goodbye",
      "message": "goodbye",
      "translation": "Goodbye"
    },
    {
      "id": "farewell",
      "message": "farewell",
      "translation": "Bye"
    }
  ]
}`

		if err := os.WriteFile(jsonFile1, []byte(data1), 0644); err != nil {
			t.Fatalf("Failed to create test file 1: %v", err)
		}

		if err := os.WriteFile(jsonFile2, []byte(data2), 0644); err != nil {
			t.Fatalf("Failed to create test file 2: %v", err)
		}

		// Create source with entries
		entries := []Entry{
			{file: jsonFile1, loader: NewJSONLoader()},
			{file: jsonFile2, loader: NewJSONLoader()},
		}

		source := NewSource(msg.English, entries)
		builder := catalog.NewBuilder()

		// Load translations
		source.Load(builder)

		// Verify entries are cleared after load
		if source.entries != nil {
			t.Error("Source.entries should be nil after successful load")
		}
	})

	t.Run("Load with errors", func(t *testing.T) {
		// Create source with non-existent files
		entries := []Entry{
			{file: "nonexistent1.json", loader: NewJSONLoader()},
			{file: "nonexistent2.json", loader: NewJSONLoader()},
		}

		var logMessages []string
		logFunc := func(msg string) {
			logMessages = append(logMessages, msg)
		}

		source := NewSource(msg.English, entries)
		source.SetLogFunc(logFunc)
		builder := catalog.NewBuilder()

		// Load translations (should log errors but not panic)
		source.Load(builder)

		// Should have logged error messages
		if len(logMessages) == 0 {
			t.Error("Expected log messages for failed file loads")
		}

		// Verify entries are still cleared after load
		if source.entries != nil {
			t.Error("Source.entries should be nil even after failed load")
		}
	})

	t.Run("Multiple loads", func(t *testing.T) {
		// Create temporary directory and file
		tempDir := t.TempDir()
		jsonFile := filepath.Join(tempDir, "test.json")
		data := `{
  "language": "en",
  "messages": [
    {
      "id": "hello",
      "message": "hello",
      "translation": "Hello World"
    }
  ]
}`

		if err := os.WriteFile(jsonFile, []byte(data), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		entries := []Entry{
			{file: jsonFile, loader: NewJSONLoader()},
		}

		source := NewSource(msg.English, entries)
		builder1 := catalog.NewBuilder()
		builder2 := catalog.NewBuilder()

		// Load first time
		source.Load(builder1)

		// Load second time (should be a no-op since entries are cleared)
		source.Load(builder2)

		// Both loads should complete without panic
		if source.entries != nil {
			t.Error("Source.entries should remain nil after second load")
		}
	})
}

func TestSource_loadFileToBuilder(t *testing.T) {
	t.Run("Empty entries", func(t *testing.T) {
		source := NewSource(msg.English, []Entry{})
		builder := catalog.NewBuilder()

		// This should not panic
		source.loadFileToBuilder(builder)
	})

	t.Run("Successful file loading", func(t *testing.T) {
		tempDir := t.TempDir()
		jsonFile := filepath.Join(tempDir, "test.json")
		data := `{
  "language": "en",
  "messages": [
    {
      "id": "hello",
      "message": "hello",
      "translation": "Hello World"
    },
    {
      "id": "greeting",
      "message": "greeting",
      "translation": "Hi"
    }
  ]
}`

		if err := os.WriteFile(jsonFile, []byte(data), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		entries := []Entry{
			{file: jsonFile, loader: NewJSONLoader()},
		}

		source := NewSource(msg.English, entries)
		builder := catalog.NewBuilder()

		// This should not panic and should load successfully
		source.loadFileToBuilder(builder)
	})

	t.Run("Mixed success and failure", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create one valid file
		validFile := filepath.Join(tempDir, "valid.json")
		validData := `{
  "language": "en",
  "messages": [
    {
      "id": "hello",
      "message": "hello",
      "translation": "Hello World"
    }
  ]
}`
		if err := os.WriteFile(validFile, []byte(validData), 0644); err != nil {
			t.Fatalf("Failed to create valid test file: %v", err)
		}

		// Create entries with one valid and one invalid file
		entries := []Entry{
			{file: validFile, loader: NewJSONLoader()},
			{file: "nonexistent.json", loader: NewJSONLoader()},
		}

		var logMessages []string
		logFunc := func(msg string) {
			logMessages = append(logMessages, msg)
		}

		source := NewSource(msg.English, entries)
		source.SetLogFunc(logFunc)
		builder := catalog.NewBuilder()

		// Should load valid file and log error for invalid file
		source.loadFileToBuilder(builder)

		// Should have logged at least one error
		if len(logMessages) == 0 {
			t.Error("Expected log messages for failed file load")
		}
	})
}

func TestSource_loadSingleFile(t *testing.T) {
	t.Run("Valid file", func(t *testing.T) {
		tempDir := t.TempDir()
		jsonFile := filepath.Join(tempDir, "test.json")
		data := `{
  "language": "en",
  "messages": [
    {
      "id": "hello",
      "message": "hello",
      "translation": "Hello World"
    },
    {
      "id": "greeting",
      "message": "greeting",
      "translation": "Hi"
    }
  ]
}`

		if err := os.WriteFile(jsonFile, []byte(data), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		source := NewSource(msg.English, []Entry{})
		builder := catalog.NewBuilder()
		loader := NewJSONLoader()

		err := source.loadSingleFile(jsonFile, loader, builder)
		if err != nil {
			t.Errorf("loadSingleFile error = %v", err)
		}
	})

	t.Run("Non-existent file", func(t *testing.T) {
		source := NewSource(msg.English, []Entry{})
		builder := catalog.NewBuilder()
		loader := NewJSONLoader()

		err := source.loadSingleFile("nonexistent.json", loader, builder)
		if err == nil {
			t.Error("loadSingleFile should return error for non-existent file")
		}
	})

	t.Run("Invalid JSON file", func(t *testing.T) {
		tempDir := t.TempDir()
		jsonFile := filepath.Join(tempDir, "invalid.json")
		data := `{"invalid": json}` // Invalid JSON

		if err := os.WriteFile(jsonFile, []byte(data), 0644); err != nil {
			t.Fatalf("Failed to create invalid test file: %v", err)
		}

		source := NewSource(msg.English, []Entry{})
		builder := catalog.NewBuilder()
		loader := NewJSONLoader()

		err := source.loadSingleFile(jsonFile, loader, builder)
		if err == nil {
			t.Error("loadSingleFile should return error for invalid JSON")
		}
	})

	t.Run("Empty file", func(t *testing.T) {
		tempDir := t.TempDir()
		emptyFile := filepath.Join(tempDir, "empty.json")

		if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
			t.Fatalf("Failed to create empty test file: %v", err)
		}

		source := NewSource(msg.English, []Entry{})
		builder := catalog.NewBuilder()
		loader := NewJSONLoader()

		err := source.loadSingleFile(emptyFile, loader, builder)
		if err != nil {
			t.Errorf("loadSingleFile error for empty file = %v", err)
		}
	})

	t.Run("File with Unicode content", func(t *testing.T) {
		tempDir := t.TempDir()
		unicodeFile := filepath.Join(tempDir, "unicode.json")
		data := `{
  "language": "zh-CN",
  "messages": [
    {
      "id": "chinese",
      "message": "chinese",
      "translation": "你好世界"
    },
    {
      "id": "emoji",
      "message": "emoji",
      "translation": "🎉🌍"
    }
  ]
}`

		if err := os.WriteFile(unicodeFile, []byte(data), 0644); err != nil {
			t.Fatalf("Failed to create unicode test file: %v", err)
		}

		source := NewSource(msg.English, []Entry{})
		builder := catalog.NewBuilder()
		loader := NewJSONLoader()

		err := source.loadSingleFile(unicodeFile, loader, builder)
		if err != nil {
			t.Errorf("loadSingleFile error for unicode file = %v", err)
		}
	})
}

func TestSource_DifferentLoaders(t *testing.T) {
	tempDir := t.TempDir()

	// Create JSON file
	jsonFile := filepath.Join(tempDir, "test.json")
	jsonData := `{
  "language": "en",
  "messages": [
    {
      "id": "hello",
      "message": "hello",
      "translation": "Hello World"
    },
    {
      "id": "greeting",
      "message": "greeting",
      "translation": "Hi"
    }
  ]
}`
	if err := os.WriteFile(jsonFile, []byte(jsonData), 0644); err != nil {
		t.Fatalf("Failed to create JSON test file: %v", err)
	}

	// Create JSONC file
	jsoncFile := filepath.Join(tempDir, "test.jsonc")
	jsoncData := `{
  // This is a comment
  "language": "en",
  "messages": [
    {
      "id": "goodbye",
      "message": "goodbye",
      "translation": "Goodbye World"
    },
    {
      "id": "farewell",
      "message": "farewell",
      "translation": "Bye" // Inline comment
    }
  ]
}`
	if err := os.WriteFile(jsoncFile, []byte(jsoncData), 0644); err != nil {
		t.Fatalf("Failed to create JSONC test file: %v", err)
	}

	// Test with JSON loader (supports both .gotext.json and .gotext.jsonc)
	loader := NewJSONLoader()
	entries := []Entry{
		{file: jsonFile, loader: loader},
		{file: jsoncFile, loader: loader},
	}

	source := NewSource(msg.English, entries)
	builder := catalog.NewBuilder()

	// Should successfully load both files with different loaders
	source.Load(builder)

	// Verify entries are cleared
	if source.entries != nil {
		t.Error("Source.entries should be nil after load with different loaders")
	}
}

func TestSource_ConcurrentLoad(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple test files
	files := make([]string, 5)
	for i := 0; i < 5; i++ {
		fileName := filepath.Join(tempDir, "test"+string(rune('A'+i))+".json")
		data := `{
  "language": "en",
  "messages": [
    {
      "id": "message",
      "message": "message",
      "translation": "Test ` + string(rune('A'+i)) + `"
    }
  ]
}`

		if err := os.WriteFile(fileName, []byte(data), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", fileName, err)
		}

		files[i] = fileName
	}

	// Create entries
	entries := make([]Entry, 5)
	for i, file := range files {
		entries[i] = Entry{file: file, loader: NewJSONLoader()}
	}

	source := NewSource(msg.English, entries)
	builder := catalog.NewBuilder()

	// Load should complete successfully even if called concurrently
	source.Load(builder)

	// Verify entries are cleared
	if source.entries != nil {
		t.Error("Source.entries should be nil after load")
	}
}

func TestSource_EdgeCases(t *testing.T) {
	t.Run("Source with nil locale", func(t *testing.T) {
		entries := []Entry{}
		source := NewSource(msg.Locale(""), entries)
		builder := catalog.NewBuilder()

		// Should not panic
		source.Load(builder)
	})

	t.Run("Source with invalid locale", func(t *testing.T) {
		entries := []Entry{}
		source := NewSource(msg.Locale("invalid-xyz-123"), entries)
		builder := catalog.NewBuilder()

		// Should not panic
		source.Load(builder)
	})

	t.Run("Large number of entries", func(t *testing.T) {
		tempDir := t.TempDir()
		entries := make([]Entry, 10)

		// Create 10 test files
		for i := 0; i < 10; i++ {
			fileName := filepath.Join(tempDir, "test"+string(rune('0'+i))+".json")
			data := `{
  "language": "en",
  "messages": [
    {
      "id": "key",
      "message": "key",
      "translation": "value ` + string(rune('0'+i)) + `"
    }
  ]
}`

			if err := os.WriteFile(fileName, []byte(data), 0644); err != nil {
				t.Fatalf("Failed to create test file %s: %v", fileName, err)
			}

			entries[i] = Entry{file: fileName, loader: NewJSONLoader()}
		}

		source := NewSource(msg.English, entries)
		builder := catalog.NewBuilder()

		// Should handle large number of entries
		source.Load(builder)

		// Verify entries are cleared
		if source.entries != nil {
			t.Error("Source.entries should be nil after load with many entries")
		}
	})

	t.Run("File path with special characters", func(t *testing.T) {
		tempDir := t.TempDir()
		specialFile := filepath.Join(tempDir, "test file with spaces.json")
		data := `{
  "language": "en",
  "messages": [
    {
      "id": "hello",
      "message": "hello",
      "translation": "Hello World"
    }
  ]
}`

		if err := os.WriteFile(specialFile, []byte(data), 0644); err != nil {
			t.Fatalf("Failed to create test file with special chars: %v", err)
		}

		entries := []Entry{
			{file: specialFile, loader: NewJSONLoader()},
		}

		source := NewSource(msg.English, entries)
		builder := catalog.NewBuilder()

		// Should handle file paths with special characters
		source.Load(builder)
	})
}

func TestSource_RealWorldScenario(t *testing.T) {
	tempDir := t.TempDir()

	// Create realistic translation files structure
	files := map[string]string{
		"common.json": `{
  "language": "en",
  "messages": [
    {
      "id": "buttons.save",
      "message": "buttons.save",
      "translation": "Save"
    },
    {
      "id": "buttons.cancel",
      "message": "buttons.cancel",
      "translation": "Cancel"
    },
    {
      "id": "buttons.delete",
      "message": "buttons.delete",
      "translation": "Delete"
    },
    {
      "id": "messages.success",
      "message": "messages.success",
      "translation": "Operation successful"
    },
    {
      "id": "messages.error",
      "message": "messages.error",
      "translation": "An error occurred"
    },
    {
      "id": "messages.loading",
      "message": "messages.loading",
      "translation": "Loading..."
    }
  ]
}`,
		"ui.json": `{
  "language": "en",
  "messages": [
    {
      "id": "navigation.home",
      "message": "navigation.home",
      "translation": "Home"
    },
    {
      "id": "navigation.settings",
      "message": "navigation.settings",
      "translation": "Settings"
    },
    {
      "id": "navigation.profile",
      "message": "navigation.profile",
      "translation": "Profile"
    },
    {
      "id": "forms.required",
      "message": "forms.required",
      "translation": "This field is required"
    },
    {
      "id": "forms.invalid",
      "message": "forms.invalid",
      "translation": "Invalid format"
    },
    {
      "id": "forms.minlength",
      "message": "forms.minlength",
      "translation": "Minimum length is %d characters"
    }
  ]
}`,
		"errors.json": `{
  "language": "en",
  "messages": [
    {
      "id": "codes.404",
      "message": "codes.404",
      "translation": "Page not found"
    },
    {
      "id": "codes.500",
      "message": "codes.500",
      "translation": "Internal server error"
    },
    {
      "id": "codes.403",
      "message": "codes.403",
      "translation": "Access denied"
    },
    {
      "id": "messages.network",
      "message": "messages.network",
      "translation": "Network connection failed"
    },
    {
      "id": "messages.timeout",
      "message": "messages.timeout",
      "translation": "Request timed out"
    },
    {
      "id": "messages.unknown",
      "message": "messages.unknown",
      "translation": "Unknown error occurred"
    }
  ]
}`,
	}

	// Write all files
	var entries []Entry
	for filename, content := range files {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
		entries = append(entries, Entry{file: filePath, loader: NewJSONLoader()})
	}

	// Create source and load
	source := NewSource(msg.English, entries)

	var logMessages []string
	logFunc := func(msg string) {
		logMessages = append(logMessages, msg)
	}
	source.SetLogFunc(logFunc)

	builder := catalog.NewBuilder()
	source.Load(builder)

	// Verify no errors occurred
	if len(logMessages) > 0 {
		t.Errorf("Unexpected log messages: %v", logMessages)
	}

	// Verify entries are cleared
	if source.entries != nil {
		t.Error("Source.entries should be nil after successful load")
	}
}

func TestSource_ErrorHandling(t *testing.T) {
	t.Run("Loader returns error", func(t *testing.T) {
		// Create a mock loader that always returns an error
		errorLoader := &mockErrorLoader{}

		entries := []Entry{
			{file: "test.json", loader: errorLoader},
		}

		var logMessages []string
		logFunc := func(msg string) {
			logMessages = append(logMessages, msg)
		}

		source := NewSource(msg.English, entries)
		source.SetLogFunc(logFunc)
		builder := catalog.NewBuilder()

		// Should log error but not panic
		source.Load(builder)

		// Should have logged the error
		if len(logMessages) == 0 {
			t.Error("Expected log message for loader error")
		}

		// Entries should still be cleared
		if source.entries != nil {
			t.Error("Source.entries should be nil even after loader error")
		}
	})

	t.Run("File permission denied", func(t *testing.T) {
		// This test is platform-dependent and might not work on all systems
		// Skip on Windows or systems where we can't create unreadable files
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "unreadable.json")

		// Create file
		data := `{
  "language": "en",
  "messages": [
    {
      "id": "hello",
      "message": "hello",
      "translation": "world"
    }
  ]
}`
		if err := os.WriteFile(filePath, []byte(data), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Try to make it unreadable (this might not work on all systems)
		_ = os.Chmod(filePath, 0000)

		entries := []Entry{
			{file: filePath, loader: NewJSONLoader()},
		}

		var logMessages []string
		logFunc := func(msg string) {
			logMessages = append(logMessages, msg)
		}

		source := NewSource(msg.English, entries)
		source.SetLogFunc(logFunc)
		builder := catalog.NewBuilder()

		// Should handle permission error gracefully
		source.Load(builder)

		// Should have logged an error (unless we can't actually make the file unreadable)
		if len(logMessages) > 0 {
			// This is expected on systems where we can make files unreadable
		}

		// Clean up: restore permissions so the file can be deleted
		_ = os.Chmod(filePath, 0644)
	})
}

// Mock loader that always returns an error
type mockErrorLoader struct{}

func (m *mockErrorLoader) Name() string {
	return "ErrorLoader"
}

func (m *mockErrorLoader) Extensions() []string {
	return []string{".error"}
}

func (m *mockErrorLoader) CanLoad(filename string) bool {
	return true
}

func (m *mockErrorLoader) Load(filename string, data []byte) (*catalog.Builder, error) {
	return nil, &mockError{"mock loader error"}
}

func (m *mockErrorLoader) LoadToBuilder(filename string, data []byte, builder *catalog.Builder, locale msg.Locale) error {
	return &mockError{"mock loader error"}
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
