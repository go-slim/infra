# Infra

[简体中文](README.md) | English

Go infrastructure library.

## Overview

This is a Go module infrastructure library that provides common utilities and tools for Go applications.

## Current Status

📦 **Initial Setup**

This project is currently in the initial setup phase with:

- Go module configuration (`go.mod`)
- Project documentation
- Development environment setup with Jujutsu (jj)

## Development

This project uses [Jujutsu (jj)](https://github.com/martinvonz/jj) for version control.

### Getting Started

```bash
# Clone the repository
git clone <repository-url>
cd infra

# Install jj (if not already installed)
# macOS: brew install jujutsu
# Other platforms: https://github.com/martinvonz/jj/releases

# Check repository status
jj status
jj log
```

### Contributing

1. Create a new change: `jj new`
2. Make your changes
3. Run tests: `go test ./...`
4. Commit: `jj commit -m "description"`
5. Push changes to remote

## Commit Message Format

Please follow the following commit message format:

### Format

```
<type>: <description>

[Optional detailed description]

[Optional closing issues]
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code formatting (no functional changes)
- `refactor`: Code refactoring
- `test`: Adding or modifying tests
- `build`: Build system or dependencies
- `ci`: CI configuration

### Example

```
feat: add retry mechanism

Implemented exponential backoff retry algorithm with support for:
- Maximum retry count configuration
- Initial delay settings
- Maximum delay limits

Closes #1
```

## Module Information

- **Module**: `go-slim.dev/infra`
- **Go Version**: 1.25

## License

[Add your license here]
