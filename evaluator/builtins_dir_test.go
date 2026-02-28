package evaluator

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/bark-lang/barki/object"
)

func TestDirListBuiltin(t *testing.T) {
	// Create temp directory with some files for testing
	tmpDir, err := os.MkdirTemp("", "bark-dir-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test files
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)
	_ = os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	// List directory - use capture() to extract result from tuple
	input := `dir.list("` + tmpDir + `") > capture(e, r)
r > len()`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 3) // 2 files + 1 directory

	// Check that result is an array (list of filenames)
	input2 := `dir.list("` + tmpDir + `") > capture(e, r)
r`
	evaluated2 := testEval(input2)
	arr, ok := evaluated2.(*object.Array)
	if !ok {
		t.Errorf("expected Array object, got=%T (%+v)", evaluated2, evaluated2)
		return
	}

	// Verify we got string elements
	for i, elem := range arr.Elements {
		if _, ok := elem.(*object.String); !ok {
			t.Errorf("element %d is not String, got=%T", i, elem)
		}
	}
}

func TestDirListSuccess(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "bark-dir-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test that successful list returns empty map as error
	// After capture(e, r), e is bound to error and r to result
	input := `dir.list("` + tmpDir + `") > capture(e, r)
e`
	evaluated := testEval(input)

	mapObj, ok := evaluated.(*object.Map)
	if !ok {
		t.Errorf("expected Map object (empty error), got=%T (%+v)", evaluated, evaluated)
		return
	}

	if len(mapObj.Pairs) != 0 {
		t.Errorf("expected empty map (no error), got map with %d pairs", len(mapObj.Pairs))
	}
}

func TestDirListErrors(t *testing.T) {
	// Try to list non-existent directory
	input := `dir.list("/nonexistent/directory/path")`
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

	// Second element should be empty array
	emptyArr, ok := tpl.Elements[1].(*object.Array)
	if !ok {
		t.Errorf("expected Array at index 1, got=%T", tpl.Elements[1])
	} else if len(emptyArr.Elements) != 0 {
		t.Errorf("expected empty array at index 1, got %d elements", len(emptyArr.Elements))
	}
}

func TestDirExistsBuiltin(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "bark-dir-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Also create a file to test the difference
	testFile := filepath.Join(tmpDir, "file.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)

	// Directory exists
	input1 := `dir.exists?("` + tmpDir + `")`
	evaluated := testEval(input1)
	testBooleanObject(t, evaluated, true)

	// Directory doesn't exist
	input2 := `dir.exists?("` + filepath.Join(tmpDir, "nonexistent") + `")`
	evaluated = testEval(input2)
	testBooleanObject(t, evaluated, false)

	// File is not a directory
	input3 := `dir.exists?("` + testFile + `")`
	evaluated = testEval(input3)
	testBooleanObject(t, evaluated, false)
}

func TestDirAbsentBuiltin(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "bark-dir-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Directory exists, so absent? is false
	input1 := `dir.absent?("` + tmpDir + `")`
	evaluated := testEval(input1)
	testBooleanObject(t, evaluated, false)

	// Directory doesn't exist, so absent? is true
	input2 := `dir.absent?("` + filepath.Join(tmpDir, "nonexistent") + `")`
	evaluated = testEval(input2)
	testBooleanObject(t, evaluated, true)
}

func TestDirWrongArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// dir.list errors
		{`dir.list()`, "dir.list requires 1 argument (path), got=0"},
		{`dir.list("a", "b")`, "dir.list requires 1 argument (path), got=2"},
		{`dir.list(123)`, "dir.list requires string argument, got=INTEGER"},

		// dir.exists? errors
		{`dir.exists?()`, "dir.exists? requires 1 argument (path), got=0"},
		{`dir.exists?("a", "b")`, "dir.exists? requires 1 argument (path), got=2"},
		{`dir.exists?(123)`, "dir.exists? requires string argument, got=INTEGER"},

		// dir.absent? errors
		{`dir.absent?()`, "dir.absent? requires 1 argument (path), got=0"},
		{`dir.absent?("a", "b")`, "dir.absent? requires 1 argument (path), got=2"},
		{`dir.absent?(123)`, "dir.absent? requires string argument, got=INTEGER"},
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
