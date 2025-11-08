# Message Internationalization and Localization

[![Go Reference](https://pkg.go.dev/badge/go-slim.dev/infra/msg.svg)](https://pkg.go.dev/go-slim.dev/infra/msg)
[![Go Report Card](https://goreportcard.com/badge/go-slim.dev/infra/msg)](https://goreportcard.com/report/go-slim.dev/infra/msg)
[![Test Status](https://github.com/go-slim/msg/workflows/Test/badge.svg)](https://github.com/go-slim/msg/actions?query=workflow%3ATest)

A comprehensive internationalization (i18n) and localization (l10n) package for Go applications, providing robust support for message formatting, pluralization, and locale management.

## Features

- 🏷️ **BCP 47 Compliant**: Full support for language tags and locale identifiers
- 🌍 **Context-Aware**: Seamless integration with `context.Context` for request-scoped localization
- 🧩 **Extensible**: Pluggable printer factories and formatters
- 🚀 **High Performance**: Efficient caching and zero-allocation design
- 🛡️ **Thread-Safe**: Safe for concurrent use across goroutines
- 📚 **Rich Locale Support**: Built-in support for common languages and regions

## Installation

```bash
go get go-slim.dev/infra/msg
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"go-slim.dev/infra/msg"
)

func main() {
	// Create a context with locale
	ctx := msg.WithLocaleContext(context.Background(), msg.ChineseSimplified)

	// Get the locale from context
	if locale, ok := msg.GetLocaleFromContext(ctx); ok {
		fmt.Printf("Current locale: %s\n", locale)
	}

	// Example of using a simple formatter
	printer := msg.NewSimplePrinter()
	message := printer.Sprintf("Welcome to our application!")
	fmt.Println(message)
}
```

## Core Concepts

### Locales

Locales represent language and region combinations using BCP 47 format (e.g., "en-US", "zh-Hans-CN"). The package provides:

- Predefined common locales
- Parsing and validation
- Language, script, and region extraction
- Locale matching and fallback

### Printers

Printers handle the actual message formatting. The package includes:

- Basic string formatting
- Number and currency formatting
- Date and time formatting
- Pluralization and gender support

### Context Integration

Seamlessly manage locale and printer instances using Go's context:

```go
// Set locale in context
ctx := msg.WithLocaleContext(context.Background(), msg.EnglishUS)

// Get locale from context
if locale, ok := msg.GetLocaleFromContext(ctx); ok {
    // Use the locale
}
```

## Advanced Usage

### Using xtext Package

The `xtext` package provides a powerful implementation of the `PrinterFactory` interface using `golang.org/x/text` for internationalization and localization.

#### Basic Usage

```go
package main

import (
	"context"
	"log"
	"os"

	"go-slim.dev/infra/msg"
	"go-slim.dev/infra/msg/xtext"
)

func main() {
	// Create a new printer factory with configuration
	factory := xtext.NewPrinterFactory(
		xtext.BaseDir("./locales"),  // Automatically load translations from this directory
		xtext.Fallback(msg.English),  // Fallback to English if translation not found
		xtext.LogFunc(func(msg string) {
			log.Printf("[xtext] %s", msg)
		}),
	)

	// Create a context with the printer factory
	ctx := msg.WithPrinterFactoryContext(context.Background(), factory)

	// Get a printer for a specific locale
	printer, err := factory.CreatePrinter(msg.ChineseSimplified)
	if err != nil {
		log.Fatalf("Failed to create printer: %v", err)
	}
	message := printer.Sprintf("welcome_message")
	log.Println(message)
}
```

#### File Structure

```
./locales/
├── en.json
├── zh-Hans.json
└── zh-Hant.json
```

Example `en.json`:
```json
{
  "welcome_message": "Welcome to our application!",
  "user_greeting": "Hello, %s!"
}
```

#### Code Generation

For better type safety and IDE support, you can generate Go code from your translation files:

1. Create a `generate.go` file in your package:

```go
//go:generate xtext generate -pkg myapp -o messages.gen.go
```

2. Run the generator:

```bash
go generate ./...
```

This will generate strongly-typed message keys and helper functions.

### Custom Formatters

Create custom formatters by implementing the `Printer` interface:

```go
type MyFormatter struct{}

func (f *MyFormatter) Sprintf(format string, args ...interface{}) string {
    // Custom formatting logic
    return fmt.Sprintf("Formatted: "+format, args...)
}

// Create and use the formatter
printer := &MyFormatter{}
// In a real application, you would typically create a factory and set it in the context
// factory := xtext.NewPrinterFactory()
// factory.RegisterFormatter("myformat", printer)
// ctx := msg.WithPrinterFactoryContext(ctx, factory)
```

### Locale Matching

Handle locale fallback and matching:

```go
import (
    "fmt"
    "go-slim.dev/infra/msg"
)

func example() {
    supported := []msg.Locale{msg.English, msg.ChineseSimplified, msg.Spanish}
    preferred := msg.NewLocale("zh-Hans-CN")

    // Find the best match
    matched := msg.Match(preferred, supported...)
    fmt.Printf("Best match for %s: %s\n", preferred, matched)
}
```

## Best Practices

1. **Always use context** for locale propagation
2. **Cache formatters** when possible
3. **Validate locales** before use
4. **Handle fallbacks** for missing translations
5. **Use consistent** locale identifiers

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.