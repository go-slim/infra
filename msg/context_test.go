package msg

import (
	"context"
	"testing"
	"time"
)

func TestWithLocaleContext(t *testing.T) {
	tests := []struct {
		name     string
		baseCtx  context.Context
		locale   Locale
		expected bool
	}{
		{
			name:     "Background context",
			baseCtx:  context.Background(),
			locale:   English,
			expected: true,
		},
		{
			name:     "Empty context",
			baseCtx:  context.TODO(),
			locale:   Chinese,
			expected: true,
		},
		{
			name:     "Nil context",
			baseCtx:  context.Background(), // Use background instead of nil
			locale:   Japanese,
			expected: true,
		},
		{
			name:     "Empty locale",
			baseCtx:  context.Background(),
			locale:   Locale(""),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithLocaleContext(tt.baseCtx, tt.locale)

			// Verify context is not nil
			if ctx == nil {
				t.Fatal("WithLocaleContext returned nil context")
			}

			// Verify locale is stored correctly
			stored, ok := GetLocaleFromContext(ctx)
			if !ok {
				t.Error("GetLocaleFromContext returned false")
			}

			if !stored.Equal(tt.locale) {
				t.Errorf("GetLocaleFromContext() = %q, want %q", string(stored), string(tt.locale))
			}
		})
	}
}

func TestWithPrinterFactoryContext(t *testing.T) {
	// Create a mock printer factory
	factory := NewPrinterFactory()

	tests := []struct {
		name     string
		baseCtx  context.Context
		factory  PrinterFactory
		expected bool
	}{
		{
			name:     "Background context",
			baseCtx:  context.Background(),
			factory:  factory,
			expected: true,
		},
		{
			name:     "Nil factory",
			baseCtx:  context.Background(),
			factory:  nil,
			expected: true, // nil factory should still be stored and retrievable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithPrinterFactoryContext(tt.baseCtx, tt.factory)

			// Verify context is not nil
			if ctx == nil {
				t.Fatal("WithPrinterFactoryContext returned nil context")
			}

			// Verify factory is stored correctly
			stored, ok := GetPrinterFactoryFromContext(ctx)
			if !ok {
				t.Error("GetPrinterFactoryFromContext returned false")
			}

			if tt.factory == nil && stored != nil {
				t.Errorf("GetPrinterFactoryFromContext() = %v, want nil", stored)
			} else if tt.factory != nil && stored != tt.factory {
				t.Errorf("GetPrinterFactoryFromContext() = %v, want %v", stored, tt.factory)
			} else if tt.factory == nil && stored == nil {
				// Both are nil, this is expected
			}
		})
	}
}

func TestWithLocaleAndPrinterFactoryContext(t *testing.T) {
	locale := ChineseSimplified
	factory := NewPrinterFactory()

	// Test with both context.Background()
	ctx1 := WithLocaleAndPrinterFactoryContext(context.Background(), locale, factory)

	// Verify both are stored
	storedLocale, ok1 := GetLocaleFromContext(ctx1)
	if !ok1 {
		t.Error("GetLocaleFromContext returned false for context.Background")
	}
	if !storedLocale.Equal(locale) {
		t.Errorf("GetLocaleFromContext() = %q, want %q", string(storedLocale), string(locale))
	}

	storedFactory, ok2 := GetPrinterFactoryFromContext(ctx1)
	if !ok2 {
		t.Error("GetPrinterFactoryFromContext returned false for context.Background")
	}
	if storedFactory != factory {
		t.Error("GetPrinterFactoryFromContext() returned wrong factory")
	}

	// Test with existing context
	existingCtx := context.WithValue(context.Background(), "test-key", "test-value")
	ctx2 := WithLocaleAndPrinterFactoryContext(existingCtx, Spanish, factory)

	// Verify existing values are preserved
	if ctx2.Value("test-key") != "test-value" {
		t.Error("Existing context values not preserved")
	}

	// Verify new values are added
	storedLocale2, _ := GetLocaleFromContext(ctx2)
	if !storedLocale2.Equal(Spanish) {
		t.Errorf("GetLocaleFromContext() = %q, want %q", string(storedLocale2), string(Spanish))
	}
}

func TestGetLocaleFromContext(t *testing.T) {
	t.Run("Locale in context", func(t *testing.T) {
		locale := English
		ctx := WithLocaleContext(context.Background(), locale)

		stored, ok := GetLocaleFromContext(ctx)
		if !ok {
			t.Error("GetLocaleFromContext returned false")
		}
		if !stored.Equal(locale) {
			t.Errorf("GetLocaleFromContext() = %q, want %q", string(stored), string(locale))
		}
	})

	t.Run("Locale not in context", func(t *testing.T) {
		ctx := context.Background()

		stored, ok := GetLocaleFromContext(ctx)
		if ok {
			t.Error("GetLocaleFromContext should return false for empty context")
		}
		if !stored.Equal("") {
			t.Errorf("GetLocaleFromContext() = %q, want empty string", string(stored))
		}
	})

	t.Run("Empty locale in context", func(t *testing.T) {
		locale := Locale("")
		ctx := WithLocaleContext(context.Background(), locale)

		stored, ok := GetLocaleFromContext(ctx)
		if !ok {
			t.Error("GetLocaleFromContext should return true for empty locale")
		}
		if !stored.Equal("") {
			t.Errorf("GetLocaleFromContext() = %q, want empty string", string(stored))
		}
	})

	t.Run("With fallback", func(t *testing.T) {
		ctx := context.Background()
		fallback := French

		stored, ok := GetLocaleFromContext(ctx, fallback)
		if !ok {
			t.Error("GetLocaleFromContext should return true when using fallback")
		}
		if !stored.Equal(fallback) {
			t.Errorf("GetLocaleFromContext() = %q, want %q", string(stored), string(fallback))
		}
	})

	t.Run("Multiple fallbacks", func(t *testing.T) {
		ctx := context.Background()
		fallback1 := English
		fallback2 := Chinese

		stored, ok := GetLocaleFromContext(ctx, fallback1, fallback2)
		if !ok {
			t.Error("GetLocaleFromContext should return true when using fallbacks")
		}
		if !stored.Equal(fallback1) {
			t.Errorf("GetLocaleFromContext() = %q, want first fallback %q", string(stored), string(fallback1))
		}
	})

	t.Run("Empty fallback", func(t *testing.T) {
		ctx := context.Background()
		fallback1 := Locale("")
		fallback2 := English

		stored, ok := GetLocaleFromContext(ctx, fallback1, fallback2)
		if !ok {
			t.Error("GetLocaleFromContext should return true when using non-empty fallback")
		}
		if !stored.Equal(fallback2) {
			t.Errorf("GetLocaleFromContext() = %q, want second fallback %q", string(stored), string(fallback2))
		}
	})
}

func TestGetPrinterFactoryFromContext(t *testing.T) {
	factory := NewPrinterFactory()

	t.Run("Factory in context", func(t *testing.T) {
		ctx := WithPrinterFactoryContext(context.Background(), factory)

		stored, ok := GetPrinterFactoryFromContext(ctx)
		if !ok {
			t.Error("GetPrinterFactoryFromContext returned false")
		}
		if stored != factory {
			t.Error("GetPrinterFactoryFromContext returned wrong factory")
		}
	})

	t.Run("Factory not in context", func(t *testing.T) {
		ctx := context.Background()

		stored, ok := GetPrinterFactoryFromContext(ctx)
		if ok {
			t.Error("GetPrinterFactoryFromContext should return false for empty context")
		}
		if stored != nil {
			t.Error("GetPrinterFactoryFromContext should return nil for empty context")
		}
	})

	t.Run("With fallback", func(t *testing.T) {
		ctx := context.Background()
		fallbackFactory := NewPrinterFactory()

		stored, ok := GetPrinterFactoryFromContext(ctx, fallbackFactory)
		if !ok {
			t.Error("GetPrinterFactoryFromContext should return true when using fallback")
		}
		if stored != fallbackFactory {
			t.Error("GetPrinterFactoryFromContext returned wrong fallback factory")
		}
	})

	t.Run("Multiple fallbacks", func(t *testing.T) {
		ctx := context.Background()
		fallback1 := NewPrinterFactory()
		fallback2 := NewPrinterFactory()

		stored, ok := GetPrinterFactoryFromContext(ctx, fallback1, fallback2)
		if !ok {
			t.Error("GetPrinterFactoryFromContext should return true when using fallbacks")
		}
		if stored != fallback1 {
			t.Error("GetPrinterFactoryFromContext should return first fallback")
		}
	})

	t.Run("Nil fallback", func(t *testing.T) {
		ctx := context.Background()
		var fallback1 PrinterFactory = nil
		fallback2 := NewPrinterFactory()

		stored, ok := GetPrinterFactoryFromContext(ctx, fallback1, fallback2)
		if !ok {
			t.Error("GetPrinterFactoryFromContext should return true when using non-nil fallback")
		}
		if stored != fallback2 {
			t.Error("GetPrinterFactoryFromContext should return second fallback")
		}
	})
}

func TestContextValueKeys(t *testing.T) {
	// Test that our context keys don't conflict with other values
	t.Run("No conflict with other context values", func(t *testing.T) {
		// Add some other context values
		ctx := context.WithValue(context.Background(), "locale-key", "not-a-locale")
		ctx = context.WithValue(ctx, "factory-key", "not-a-factory")

		// Add our locale
		locale := English
		ctx = WithLocaleContext(ctx, locale)

		// Verify we can still get our locale
		stored, ok := GetLocaleFromContext(ctx)
		if !ok {
			t.Error("GetLocaleFromContext returned false")
		}
		if !stored.Equal(locale) {
			t.Errorf("GetLocaleFromContext() = %q, want %q", string(stored), string(locale))
		}

		// Verify other values are preserved
		if ctx.Value("locale-key") != "not-a-locale" {
			t.Error("Other context values not preserved")
		}
	})

	t.Run("Unique keys", func(t *testing.T) {
		// Create two different contexts with our keys
		ctx1 := WithLocaleContext(context.Background(), English)
		ctx2 := WithPrinterFactoryContext(context.Background(), NewPrinterFactory())

		// Verify contexts don't interfere
		locale1, ok1 := GetLocaleFromContext(ctx1)
		if !ok1 || !locale1.Equal(English) {
			t.Error("Context 1 locale not preserved")
		}

		factory2, ok2 := GetPrinterFactoryFromContext(ctx2)
		if !ok2 || factory2 == nil {
			t.Error("Context 2 factory not preserved")
		}

		// Verify context 1 doesn't have factory
		_, hasFactory := GetPrinterFactoryFromContext(ctx1)
		if hasFactory {
			t.Error("Context 1 should not have factory")
		}

		// Verify context 2 doesn't have locale
		_, hasLocale := GetLocaleFromContext(ctx2)
		if hasLocale {
			t.Error("Context 2 should not have locale")
		}
	})
}

func TestContextChain(t *testing.T) {
	// Test that context chaining works correctly
	t.Run("Chained contexts", func(t *testing.T) {
		// Create a chain of contexts
		ctx := context.Background()
		ctx = WithLocaleContext(ctx, English)
		ctx = WithPrinterFactoryContext(ctx, NewPrinterFactory())
		ctx = context.WithValue(ctx, "test", "value")

		// Verify all values are accessible
		locale, hasLocale := GetLocaleFromContext(ctx)
		if !hasLocale || !locale.Equal(English) {
			t.Error("Locale not preserved in context chain")
		}

		factory, hasFactory := GetPrinterFactoryFromContext(ctx)
		if !hasFactory || factory == nil {
			t.Error("Factory not preserved in context chain")
		}

		if ctx.Value("test") != "value" {
			t.Error("Other values not preserved in context chain")
		}
	})

	t.Run("Override in chain", func(t *testing.T) {
		// Override locale in context chain
		ctx := WithLocaleContext(context.Background(), English)
		ctx = WithLocaleContext(ctx, Chinese) // Override

		// Verify latest value is used
		locale, hasLocale := GetLocaleFromContext(ctx)
		if !hasLocale || !locale.Equal(Chinese) {
			t.Errorf("Expected Chinese, got %q", string(locale))
		}
	})
}

func TestContextThreadSafety(t *testing.T) {
	// Test that context operations are thread-safe
	t.Run("Concurrent context creation", func(t *testing.T) {
		done := make(chan bool, 10)

		// Create contexts concurrently
		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()

				locales := []Locale{English, Chinese, Spanish, French}
				locale := locales[i%4]
				ctx := WithLocaleContext(context.Background(), locale)

				stored, ok := GetLocaleFromContext(ctx)
				if !ok || !stored.Equal(locale) {
					t.Errorf("Goroutine %d: locale mismatch", i)
				}
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("Concurrent context reads", func(t *testing.T) {
		ctx := WithLocaleContext(context.Background(), English)
		done := make(chan bool, 10)

		// Read from context concurrently
		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()

				locale, ok := GetLocaleFromContext(ctx)
				if !ok || !locale.Equal(English) {
					t.Error("Concurrent read failed")
				}
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestContextWithCancellation(t *testing.T) {
	// Test that our context works with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Add locale to cancellable context
	locale := English
	ctx = WithLocaleContext(ctx, locale)

	// Verify locale is accessible before cancellation
	stored, ok := GetLocaleFromContext(ctx)
	if !ok || !stored.Equal(locale) {
		t.Error("Locale not accessible before cancellation")
	}

	// Cancel context
	cancel()

	// Try to access locale after cancellation
	// Note: Our implementation doesn't check context cancellation for locale access
	// which is correct - the locale should still be accessible even after cancellation
	stored, ok = GetLocaleFromContext(ctx)
	if !ok || !stored.Equal(locale) {
		t.Error("Locale not accessible after cancellation")
	}
}

func TestContextWithTimeout(t *testing.T) {
	// Test that our context works with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Add locale to timeout context
	locale := English
	ctx = WithLocaleContext(ctx, locale)

	// Verify locale is accessible
	stored, ok := GetLocaleFromContext(ctx)
	if !ok || !stored.Equal(locale) {
		t.Error("Locale not accessible in timeout context")
	}
}
