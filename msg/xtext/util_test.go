package xtext

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDirectoryForEntries(t *testing.T) {
	t.Run("Empty directory", func(t *testing.T) {
		tempDir := t.TempDir()

		registry := NewLoaderRegistry()
		entries := ScanDirectoryForEntries(tempDir, registry)

		if len(entries) != 0 {
			t.Errorf("ScanDirectoryForEntries() on empty directory = %d entries, want 0", len(entries))
		}
	})

	t.Run("Directory with gotext files", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create test files
		files := map[string]string{
			"test1.gotext.json": `{
  "language": "zh-CN",
  "messages": [
    {
      "id": "Hello",
      "message": "Hello",
      "translation": "你好"
    }
  ]
}`,
			"test2.gotext.jsonc": `{
  "language": "zh-CN",
  "messages": [
    {
      "id": "Goodbye",
      "message": "Goodbye",
      "translation": "再见"
    }
  ]
}`,
		}

		for filename, content := range files {
			filePath := filepath.Join(tempDir, filename)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to create test file %s: %v", filename, err)
			}
		}

		registry := NewLoaderRegistry()
		entries := ScanDirectoryForEntries(tempDir, registry)

		if len(entries) != len(files) {
			t.Errorf("ScanDirectoryForEntries() returned %d entries, want %d", len(entries), len(files))
		}
	})

	t.Run("Directory with mixed files", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create test files - mix of supported and unsupported
		files := map[string]string{
			"test.gotext.json": `{
  "language": "zh-CN",
  "messages": [
    {
      "id": "Hello",
      "message": "Hello",
      "translation": "你好"
    }
  ]
}`,
			"test.gotext.jsonc": `{
  "language": "zh-CN",
  "messages": [
    {
      "id": "Goodbye",
      "message": "Goodbye",
      "translation": "再见"
    }
  ]
}`,
			"test.json": `{"ignored": "true"}`, // Not .gotext.json
			"test.txt":  `ignored`,             // Not a gotext file
			"readme.md": `# Ignored`,           // Not a gotext file
		}

		for filename, content := range files {
			filePath := filepath.Join(tempDir, filename)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to create test file %s: %v", filename, err)
			}
		}

		registry := NewLoaderRegistry()
		entries := ScanDirectoryForEntries(tempDir, registry)

		// Should only find .gotext.json and .gotext.jsonc files
		expectedCount := 2
		if len(entries) != expectedCount {
			t.Errorf("ScanDirectoryForEntries() returned %d entries, want %d", len(entries), expectedCount)
		}
	})

	t.Run("Ignores subdirectories", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create files in main directory
		mainFiles := map[string]string{
			"main.gotext.json": `{
  "language": "zh-CN",
  "messages": [
    {
      "id": "Hello",
      "message": "Hello",
      "translation": "你好"
    }
  ]
}`,
		}

		for filename, content := range mainFiles {
			filePath := filepath.Join(tempDir, filename)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to create test file %s: %v", filename, err)
			}
		}

		// Create subdirectory with files
		subDir := filepath.Join(tempDir, "subdir")
		if err := os.Mkdir(subDir, 0755); err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}

		subFiles := map[string]string{
			"sub.gotext.json": `{
  "language": "zh-CN",
  "messages": [
    {
      "id": "Goodbye",
      "message": "Goodbye",
      "translation": "再见"
    }
  ]
}`,
		}

		for filename, content := range subFiles {
			filePath := filepath.Join(subDir, filename)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to create test file %s: %v", filename, err)
			}
		}

		registry := NewLoaderRegistry()
		entries := ScanDirectoryForEntries(tempDir, registry)

		// Should only find main directory files
		if len(entries) != len(mainFiles) {
			t.Errorf("ScanDirectoryForEntries() returned %d entries, want %d (subdirectories should be ignored)", len(entries), len(mainFiles))
		}
	})

	t.Run("Non-existent directory", func(t *testing.T) {
		nonExistentDir := "/path/that/does/not/exist"

		registry := NewLoaderRegistry()
		entries := ScanDirectoryForEntries(nonExistentDir, registry)

		// Should return empty entries, not panic
		if len(entries) != 0 {
			t.Errorf("ScanDirectoryForEntries() on non-existent directory = %d entries, want 0", len(entries))
		}
	})

	t.Run("Case insensitivity", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create files with different cases
		files := map[string]string{
			"test.GOTEXT.JSON": `{
  "language": "zh-CN",
  "messages": [
    {
      "id": "Hello",
      "message": "Hello",
      "translation": "你好"
    }
  ]
}`,
			"test.Gotext.Jsonc": `{
  "language": "zh-CN",
  "messages": [
    {
      "id": "Goodbye",
      "message": "Goodbye",
      "translation": "再见"
    }
  ]
}`,
		}

		for filename, content := range files {
			filePath := filepath.Join(tempDir, filename)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to create test file %s: %v", filename, err)
			}
		}

		registry := NewLoaderRegistry()
		entries := ScanDirectoryForEntries(tempDir, registry)

		// Should find both files (case-insensitive matching)
		if len(entries) != len(files) {
			t.Errorf("ScanDirectoryForEntries() returned %d entries, want %d", len(entries), len(files))
		}
	})

	t.Run("Empty files", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create empty files
		emptyFiles := []string{
			"empty1.gotext.json",
			"empty2.gotext.jsonc",
		}

		for _, filename := range emptyFiles {
			filePath := filepath.Join(tempDir, filename)
			if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
				t.Fatalf("Failed to create empty file %s: %v", filename, err)
			}
		}

		registry := NewLoaderRegistry()
		entries := ScanDirectoryForEntries(tempDir, registry)

		// Should still find the files (based on extension only)
		if len(entries) != len(emptyFiles) {
			t.Errorf("ScanDirectoryForEntries() returned %d entries, want %d", len(entries), len(emptyFiles))
		}
	})
}
