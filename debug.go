package main

import (
	"fmt"

	"go-slim.dev/infra/msg/xtext"
)

func main() {
	registry := xtext.NewLoaderRegistry()
	fmt.Println("Debug: Loaders and extensions:")

	// Get all loaders and their extensions
	allLoaders := registry.GetAllLoaders()
	for _, loader := range allLoaders {
		fmt.Printf("Loader %s: %v\n", loader.Name(), loader.Extensions())
	}

	// Test specific files
	testFiles := []string{
		"test.JSON",
		"test.Json",
		"test.json",
		"TEST.JSONC",
		"test.jsonc",
	}

	for _, filename := range testFiles {
		loader, ok := registry.GetLoaderForFile(filename)
		if ok {
			fmt.Printf("File: %s -> Loader found: %v (%s)\n", filename, ok, loader.Name())
		} else {
			fmt.Printf("File: %s -> Loader found: %v\n", filename, ok)
		}
	}
}
