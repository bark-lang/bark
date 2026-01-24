package evaluator

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/bark-lang/bark/object"
)

func TestFileWriteAndReadBuiltin(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "bark-file-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")

	// Write a file
	writeInput := `file.write("` + testFile + `", "hello world")`
	evaluated := testEval(writeInput)

	// Should return empty map (no error)
	mapObj, ok := evaluated.(*object.Map)
	if !ok {
		t.Errorf("expected Map object (success), got=%T (%+v)", evaluated, evaluated)
	} else if len(mapObj.Pairs) != 0 {
		t.Errorf("expected empty map (no error), got map with %d pairs", len(mapObj.Pairs))
	}

	// Read it back - use capture() to extract result from tuple
	readInput := `file.read("` + testFile + `") > capture(e, r)
r`
	evaluated = testEval(readInput)
	testStringObject(t, evaluated, "hello world")
}

func TestFileWriteCreatesDirectories(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "bark-file-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Write to nested path that doesn't exist
	nestedFile := filepath.Join(tmpDir, "subdir", "nested", "test.txt")
	writeInput := `file.write("` + nestedFile + `", "nested content")`
	evaluated := testEval(writeInput)

	// Should succeed
	if _, ok := evaluated.(*object.Map); !ok {
		t.Errorf("expected Map object (success), got=%T (%+v)", evaluated, evaluated)
	}

	// Verify directories were created
	if _, err := os.Stat(filepath.Join(tmpDir, "subdir", "nested")); os.IsNotExist(err) {
		t.Errorf("directories were not created")
	}
}

func TestFileAppendBuiltin(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "bark-file-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "append.txt")

	// Write initial content
	testEval(`file.write("` + testFile + `", "line1\n")`)

	// Append more content
	testEval(`file.append("` + testFile + `", "line2\n")`)
	testEval(`file.append("` + testFile + `", "line3")`)

	// Read it back - use capture() to extract result from tuple
	readInput := `file.read("` + testFile + `") > capture(e, r)
r`
	evaluated := testEval(readInput)
	testStringObject(t, evaluated, "line1\nline2\nline3")
}

func TestFileExistsBuiltin(t *testing.T) {
	// Create temp directory and file for testing
	tmpDir, err := os.MkdirTemp("", "bark-file-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "exists.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)

	// File exists
	input1 := `file.exists?("` + testFile + `")`
	evaluated := testEval(input1)
	testBooleanObject(t, evaluated, true)

	// File doesn't exist
	input2 := `file.exists?("` + filepath.Join(tmpDir, "nonexistent.txt") + `")`
	evaluated = testEval(input2)
	testBooleanObject(t, evaluated, false)

	// Directory is not a file
	input3 := `file.exists?("` + tmpDir + `")`
	evaluated = testEval(input3)
	testBooleanObject(t, evaluated, false)
}

func TestFileAbsentBuiltin(t *testing.T) {
	// Create temp directory and file for testing
	tmpDir, err := os.MkdirTemp("", "bark-file-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "exists.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)

	// File exists, so absent? is false
	input1 := `file.absent?("` + testFile + `")`
	evaluated := testEval(input1)
	testBooleanObject(t, evaluated, false)

	// File doesn't exist, so absent? is true
	input2 := `file.absent?("` + filepath.Join(tmpDir, "nonexistent.txt") + `")`
	evaluated = testEval(input2)
	testBooleanObject(t, evaluated, true)
}

func TestFileDeleteBuiltin(t *testing.T) {
	// Create temp directory and file for testing
	tmpDir, err := os.MkdirTemp("", "bark-file-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "delete.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)

	// File exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatalf("test file was not created")
	}

	// Delete it
	deleteInput := `file.delete("` + testFile + `")`
	evaluated := testEval(deleteInput)

	// Should return empty map (no error)
	if _, ok := evaluated.(*object.Map); !ok {
		t.Errorf("expected Map object (success), got=%T (%+v)", evaluated, evaluated)
	}

	// File should no longer exist
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Errorf("file still exists after delete")
	}
}

func TestFileInfoBuiltin(t *testing.T) {
	// Create temp directory and file for testing
	tmpDir, err := os.MkdirTemp("", "bark-file-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "info.txt")
	testContent := "test content"
	_ = os.WriteFile(testFile, []byte(testContent), 0644)

	// Get file info - check size - use capture() to extract result from tuple
	input := `file.info("` + testFile + `") > capture(e, r)
r > get("size")`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, int64(len(testContent)))

	// Get file info - check is_dir
	input2 := `file.info("` + testFile + `") > capture(e, r)
r > get("is_dir")`
	evaluated2 := testEval(input2)
	testBooleanObject(t, evaluated2, false)
}

func TestFileReadErrors(t *testing.T) {
	// Try to read non-existent file
	input := `file.read("/nonexistent/file/path.txt")`
	evaluated := testEval(input)

	// Should return error tuple
	tpl, ok := evaluated.(*object.Tuple)
	if !ok {
		t.Errorf("expected Tuple (error tuple), got=%T (%+v)", evaluated, evaluated)
		return
	}

	if len(tpl.Elements) != 2 {
		t.Errorf("expected tuple with 2 elements, got=%d", len(tpl.Elements))
		return
	}

	// First element should be Error
	if _, ok := tpl.Elements[0].(*object.Error); !ok {
		t.Errorf("expected Error at index 0, got=%T", tpl.Elements[0])
	}
}

func TestFileWrongArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// file.read errors
		{`file.read()`, "file.read requires 1 argument (path), got=0"},
		{`file.read("a", "b")`, "file.read requires 1 argument (path), got=2"},
		{`file.read(123)`, "file.read requires string argument, got=INTEGER"},

		// file.write errors
		{`file.write()`, "file.write requires 2 arguments (path, content), got=0"},
		{`file.write("path")`, "file.write requires 2 arguments (path, content), got=1"},
		{`file.write(123, "content")`, "file.write requires string path, got=INTEGER"},
		{`file.write("path", 123)`, "file.write requires string content, got=INTEGER"},

		// file.append errors
		{`file.append()`, "file.append requires 2 arguments (path, content), got=0"},
		{`file.append("path")`, "file.append requires 2 arguments (path, content), got=1"},

		// file.exists? errors
		{`file.exists?()`, "file.exists? requires 1 argument (path), got=0"},
		{`file.exists?(123)`, "file.exists? requires string argument, got=INTEGER"},

		// file.absent? errors
		{`file.absent?()`, "file.absent? requires 1 argument (path), got=0"},

		// file.delete errors
		{`file.delete()`, "file.delete requires 1 argument (path), got=0"},

		// file.info errors
		{`file.info()`, "file.info requires 1 argument (path), got=0"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("expected Error object for %q, got=%T (%+v)", tt.input, evaluated, evaluated)
			continue
		}
		if errObj.Msg != tt.expected {
			t.Errorf("wrong error message for %q. expected=%q, got=%q", tt.input, tt.expected, errObj.Msg)
		}
	}
}
