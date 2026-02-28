package evaluator

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/bark-lang/barki/lexer"
	"gitlab.com/bark-lang/barki/object"
	"gitlab.com/bark-lang/barki/parser"
)

// Helper function to create a temporary module file for testing
func createTempModule(t *testing.T, content string) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "bark-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	tmpFile := filepath.Join(tmpDir, "test_module.bark")
	err = os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("failed to write temp file: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return tmpFile, cleanup
}

func TestModulePathResolution(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "bark-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.bark")
	_ = os.WriteFile(testFile, []byte("module test"), 0644)

	tests := []struct {
		name           string
		importPath     string
		currentFile    string
		expectedError  bool
		expectedSuffix string
	}{
		{
			name:          "relative path gets .bark extension added",
			importPath:    "utils",
			currentFile:   filepath.Join(tmpDir, "file.bark"),
			expectedError: true, // File won't exist
		},
		{
			name:          "remote modules download to cache",
			importPath:    "https://example.com/module.bark",
			currentFile:   filepath.Join(tmpDir, "file.bark"),
			expectedError: true, // Will fail to download from non-existent URL
		},
		{
			name:           "existing file with .bark",
			importPath:     "test.bark",
			currentFile:    filepath.Join(tmpDir, "file.bark"),
			expectedError:  false,
			expectedSuffix: "test.bark",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absPath, err := ResolvePath(tt.importPath, tt.currentFile)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none, absPath: %s", absPath)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.expectedSuffix != "" {
					if filepath.Base(absPath) != tt.expectedSuffix {
						t.Errorf("expected path to end with %s, got: %s", tt.expectedSuffix, absPath)
					}
				}
			}
		})
	}
}

func TestModuleLoading(t *testing.T) {
	// Create a test module
	moduleContent := `module math

pub fn double(x int) {
	add(x, x) > return()
}(int)

fn add(a int, b int) {
	a > add(b) > return()
}(int)
`
	tmpFile, cleanup := createTempModule(t, moduleContent)
	defer cleanup()

	// Create a new module registry
	registry := NewModuleRegistry()

	// Load the module
	mod, err := LoadModule(tmpFile, registry)
	if err != nil {
		t.Fatalf("failed to load module: %v", err)
	}

	// Check module properties
	if mod.Name != "math" {
		t.Errorf("expected module name 'math', got '%s'", mod.Name)
	}

	if mod.Path != tmpFile {
		t.Errorf("expected module path '%s', got '%s'", tmpFile, mod.Path)
	}

	if mod.Evaluated {
		t.Errorf("module should not be evaluated yet")
	}

	// Evaluate the module
	err = EvaluateModule(mod, registry)
	if err != nil {
		t.Fatalf("failed to evaluate module: %v", err)
	}

	if !mod.Evaluated {
		t.Errorf("module should be marked as evaluated")
	}

	// Check public functions
	if !mod.IsPublicFunction("double") {
		t.Errorf("double should be a public function")
	}

	if mod.IsPublicFunction("add") {
		t.Errorf("add should not be a public function")
	}

	// Check functions exist in module environment
	if _, ok := mod.GetFunction("double"); !ok {
		t.Errorf("double function should exist in module environment")
	}

	if _, ok := mod.GetFunction("add"); !ok {
		t.Errorf("add function should exist in module environment")
	}
}

func TestCircularDependencyDetection(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "bark-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create two modules that import each other
	moduleA := filepath.Join(tmpDir, "a.bark")
	moduleB := filepath.Join(tmpDir, "b.bark")

	contentA := `module a
import "b.bark"
pub fn funcA() {}()`

	contentB := `module b
import "a.bark"
pub fn funcB() {}()`

	_ = os.WriteFile(moduleA, []byte(contentA), 0644)
	_ = os.WriteFile(moduleB, []byte(contentB), 0644)

	// Set current file for the global module registry
	moduleRegistry.SetCurrentFile(moduleA)

	// Try to evaluate module A (which imports B, which imports A)
	l := lexer.New(contentA)
	p := parser.New(l)
	program := p.ParseProgram()

	env := object.NewEnvironment()

	// This should detect circular dependency
	result := Eval(program, env)

	// Check for error
	if errObj, ok := result.(*object.Error); ok {
		// Good - we got an error
		if !contains(errObj.Msg, "circular dependency") {
			t.Errorf("expected circular dependency error, got: %s", errObj.Msg)
		}
	} else {
		t.Errorf("expected error for circular dependency, got: %T", result)
	}
}

func TestImportWithAlias(t *testing.T) {
	// Create a test module
	moduleContent := `module utilities

pub fn helper() {
	"helped" > return()
}(string)
`
	tmpFile, cleanup := createTempModule(t, moduleContent)
	defer cleanup()

	// Get just the filename for relative import
	tmpDir := filepath.Dir(tmpFile)
	moduleName := filepath.Base(tmpFile)

	// Create main file that imports with alias
	mainContent := `import "` + moduleName + `" as util

fn main() {
	util.helper() > return()
}(string)
`

	// Parse and evaluate
	l := lexer.New(mainContent)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	env := object.NewEnvironment()

	// Set current file for import resolution - use a file in the same directory
	currentFile := filepath.Join(tmpDir, "main.bark")
	moduleRegistry.SetCurrentFile(currentFile)

	// Evaluate the program
	result := Eval(program, env)

	if errObj, ok := result.(*object.Error); ok {
		t.Fatalf("evaluation error: %s", errObj.Msg)
	}

	// Call main function
	mainFn, ok := env.Get("main")
	if !ok {
		t.Fatalf("main function not found")
	}

	mainResult := applyFunction(mainFn, []object.Object{})
	if errObj, ok := mainResult.(*object.Error); ok {
		t.Fatalf("main function error: %s", errObj.Msg)
	}

	// Check result
	if str, ok := mainResult.(*object.String); ok {
		if str.Value != "helped" {
			t.Errorf("expected 'helped', got '%s'", str.Value)
		}
	} else {
		t.Errorf("expected string result, got %T", mainResult)
	}
}

func TestIncludeStatement(t *testing.T) {
	// Create a test include file
	includeContent := `fn helper(x int) {
	x > add(10) > return()
}(int)
`
	tmpFile, cleanup := createTempModule(t, includeContent)
	defer cleanup()

	// Get just the filename for relative include
	tmpDir := filepath.Dir(tmpFile)
	moduleName := filepath.Base(tmpFile)

	// Create main file that includes the helper
	mainContent := `include "` + moduleName + `"

fn main() {
	helper(5) > return()
}(int)
`

	// Parse and evaluate
	l := lexer.New(mainContent)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	env := object.NewEnvironment()

	// Set current file for include resolution - use a file in the same directory
	currentFile := filepath.Join(tmpDir, "main.bark")
	moduleRegistry.SetCurrentFile(currentFile)

	// Evaluate the program
	result := Eval(program, env)

	if errObj, ok := result.(*object.Error); ok {
		t.Fatalf("evaluation error: %s", errObj.Msg)
	}

	// Call main function
	mainFn, ok := env.Get("main")
	if !ok {
		t.Fatalf("main function not found")
	}

	// Check that helper is also in the same environment
	if _, ok := env.Get("helper"); !ok {
		t.Fatalf("helper function should be in the same environment (included)")
	}

	mainResult := applyFunction(mainFn, []object.Object{})
	if errObj, ok := mainResult.(*object.Error); ok {
		t.Fatalf("main function error: %s", errObj.Msg)
	}

	// Check result
	if num, ok := mainResult.(*object.Integer); ok {
		if num.Value != 15 {
			t.Errorf("expected 15, got %d", num.Value)
		}
	} else {
		t.Errorf("expected integer result, got %T", mainResult)
	}
}

func TestPublicPrivateVisibility(t *testing.T) {
	// Create a test module with public and private functions
	moduleContent := `module mymodule

pub fn publicFunc() {
	"public" > return()
}(string)

fn privateFunc() {
	"private" > return()
}(string)
`
	tmpFile, cleanup := createTempModule(t, moduleContent)
	defer cleanup()

	// Get just the filename for relative import
	tmpDir := filepath.Dir(tmpFile)
	moduleName := filepath.Base(tmpFile)

	// Create main file that tries to access both
	mainContent := `import "` + moduleName + `" as mymodule

fn testPublic() {
	mymodule.publicFunc() > return()
}(string)

fn testPrivate() {
	mymodule.privateFunc() > return()
}(string)
`

	// Parse and evaluate
	l := lexer.New(mainContent)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	env := object.NewEnvironment()

	// Set current file for import resolution - use a file in the same directory
	currentFile := filepath.Join(tmpDir, "main.bark")
	moduleRegistry.SetCurrentFile(currentFile)

	// Evaluate the program
	result := Eval(program, env)

	if errObj, ok := result.(*object.Error); ok {
		t.Fatalf("evaluation error: %s", errObj.Msg)
	}

	// Test public function access - should work
	testPublicFn, ok := env.Get("testPublic")
	if !ok {
		t.Fatalf("testPublic function not found")
	}

	publicResult := applyFunction(testPublicFn, []object.Object{})
	if errObj, ok := publicResult.(*object.Error); ok {
		t.Fatalf("testPublic error: %s", errObj.Msg)
	}

	if str, ok := publicResult.(*object.String); ok {
		if str.Value != "public" {
			t.Errorf("expected 'public', got '%s'", str.Value)
		}
	} else {
		t.Errorf("expected string result, got %T", publicResult)
	}

	// Test private function access - should fail
	testPrivateFn, ok := env.Get("testPrivate")
	if !ok {
		t.Fatalf("testPrivate function not found")
	}

	privateResult := applyFunction(testPrivateFn, []object.Object{})
	if errObj, ok := privateResult.(*object.Error); ok {
		// Good - we got an error for accessing private function
		if !contains(errObj.Msg, "not exported") {
			t.Errorf("expected 'not exported' error, got: %s", errObj.Msg)
		}
	} else {
		t.Errorf("expected error for accessing private function, got: %T", privateResult)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
