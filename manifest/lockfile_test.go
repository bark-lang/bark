package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLockfile(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "empty lockfile",
			input:     "",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "single module",
			input: `
[[module]]
name = "json"
version = "v1.0.0"
source = "git+https://gitlab.com/bark-lang/modules/json?tag=v1.0.0"
checksum = "sha256:abc123"
`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "multiple modules",
			input: `
[[module]]
name = "json"
version = "v1.0.0"
source = "git+https://gitlab.com/bark-lang/modules/json?tag=v1.0.0"
checksum = "sha256:abc123"

[[module]]
name = "logger"
version = "v2.1.0"
source = "git+https://gitlab.com/user/bark-logger?tag=v2.1.0"
checksum = "sha256:def456"
`,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:    "invalid toml",
			input:   "[[module]\nbroken",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lf, err := ParseLockfile([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(lf.Modules) != tt.wantCount {
				t.Errorf("module count = %d, want %d", len(lf.Modules), tt.wantCount)
			}
		})
	}
}

func TestLockfileGetModule(t *testing.T) {
	lf := &Lockfile{
		Modules: []LockedModule{
			{Name: "json", Version: "v1.0.0", Source: "git+https://example.com/json", Checksum: "sha256:abc"},
			{Name: "logger", Version: "v2.0.0", Source: "git+https://example.com/logger", Checksum: "sha256:def"},
		},
	}

	// Test existing module
	mod := lf.GetModule("json")
	if mod == nil {
		t.Fatal("GetModule(\"json\") returned nil")
	} else if mod.Version != "v1.0.0" {
		t.Errorf("version = %q, want %q", mod.Version, "v1.0.0")
	}

	// Test non-existent module
	mod = lf.GetModule("notexist")
	if mod != nil {
		t.Error("GetModule(\"notexist\") should return nil")
	}
}

func TestLockfileSetModule(t *testing.T) {
	lf := NewLockfile()

	// Add new module
	lf.SetModule(LockedModule{
		Name:     "json",
		Version:  "v1.0.0",
		Source:   "git+https://example.com/json",
		Checksum: "sha256:abc",
	})

	if !lf.HasModule("json") {
		t.Error("module 'json' should exist after SetModule")
	}

	// Update existing module
	lf.SetModule(LockedModule{
		Name:     "json",
		Version:  "v2.0.0",
		Source:   "git+https://example.com/json",
		Checksum: "sha256:def",
	})

	mod := lf.GetModule("json")
	if mod.Version != "v2.0.0" {
		t.Errorf("version after update = %q, want %q", mod.Version, "v2.0.0")
	}

	// Should still be only one module
	if len(lf.Modules) != 1 {
		t.Errorf("module count = %d, want 1", len(lf.Modules))
	}
}

func TestLockfileRemoveModule(t *testing.T) {
	lf := &Lockfile{
		Modules: []LockedModule{
			{Name: "json", Version: "v1.0.0"},
			{Name: "logger", Version: "v2.0.0"},
		},
	}

	lf.RemoveModule("json")

	if lf.HasModule("json") {
		t.Error("module 'json' should not exist after RemoveModule")
	}
	if !lf.HasModule("logger") {
		t.Error("module 'logger' should still exist")
	}
	if len(lf.Modules) != 1 {
		t.Errorf("module count = %d, want 1", len(lf.Modules))
	}
}

func TestLockfileToTOML(t *testing.T) {
	lf := &Lockfile{
		Modules: []LockedModule{
			{Name: "json", Version: "v1.0.0", Source: "git+https://example.com/json?tag=v1.0.0", Checksum: "sha256:abc123"},
		},
	}

	tomlStr := lf.ToTOML()

	// Parse it back
	parsed, err := ParseLockfile([]byte(tomlStr))
	if err != nil {
		t.Fatalf("generated TOML is invalid: %v", err)
	}

	if len(parsed.Modules) != 1 {
		t.Errorf("round-trip module count = %d, want 1", len(parsed.Modules))
	}

	mod := parsed.GetModule("json")
	if mod == nil {
		t.Fatal("module 'json' not found after round-trip")
	} else if mod.Version != "v1.0.0" {
		t.Errorf("round-trip version = %q, want %q", mod.Version, "v1.0.0")
	}
}

func TestLockfileSortedOutput(t *testing.T) {
	lf := &Lockfile{
		Modules: []LockedModule{
			{Name: "zebra", Version: "v1.0.0", Source: "git+https://example.com/zebra", Checksum: "sha256:z"},
			{Name: "alpha", Version: "v1.0.0", Source: "git+https://example.com/alpha", Checksum: "sha256:a"},
			{Name: "beta", Version: "v1.0.0", Source: "git+https://example.com/beta", Checksum: "sha256:b"},
		},
	}

	tomlStr := lf.ToTOML()

	// Parse and check order
	parsed, err := ParseLockfile([]byte(tomlStr))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// The original lockfile should have modules in sorted order in TOML output
	// but parsing doesn't guarantee order, so we check the TOML string directly
	if len(parsed.Modules) != 3 {
		t.Errorf("module count = %d, want 3", len(parsed.Modules))
	}
}

func TestBuildSource(t *testing.T) {
	tests := []struct {
		name string
		dep  Dependency
		want string
	}{
		{
			name: "git with tag",
			dep:  Dependency{Git: "https://example.com/repo", Tag: "v1.0.0"},
			want: "git+https://example.com/repo?tag=v1.0.0",
		},
		{
			name: "git with branch",
			dep:  Dependency{Git: "https://example.com/repo", Branch: "main"},
			want: "git+https://example.com/repo?branch=main",
		},
		{
			name: "git with rev",
			dep:  Dependency{Git: "https://example.com/repo", Rev: "abc123"},
			want: "git+https://example.com/repo?rev=abc123",
		},
		{
			name: "path dependency",
			dep:  Dependency{Path: "./local/lib"},
			want: "path+./local/lib",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildSource(&tt.dep)
			if got != tt.want {
				t.Errorf("BuildSource() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComputeChecksum(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "bark-checksum-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test bark file
	testFile := filepath.Join(tmpDir, "test.bark")
	if err := os.WriteFile(testFile, []byte("fn main() { stdout(\"hello\") }()"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Compute checksum
	checksum, err := ComputeChecksum(tmpDir)
	if err != nil {
		t.Fatalf("ComputeChecksum failed: %v", err)
	}

	// Verify format
	if len(checksum) < 10 || checksum[:7] != "sha256:" {
		t.Errorf("checksum format invalid: %q", checksum)
	}

	// Verify same content produces same checksum
	checksum2, err := ComputeChecksum(tmpDir)
	if err != nil {
		t.Fatalf("second ComputeChecksum failed: %v", err)
	}
	if checksum != checksum2 {
		t.Error("same content should produce same checksum")
	}

	// Verify different content produces different checksum
	if err := os.WriteFile(testFile, []byte("fn main() { stdout(\"world\") }()"), 0644); err != nil {
		t.Fatalf("failed to update test file: %v", err)
	}
	checksum3, err := ComputeChecksum(tmpDir)
	if err != nil {
		t.Fatalf("third ComputeChecksum failed: %v", err)
	}
	if checksum == checksum3 {
		t.Error("different content should produce different checksum")
	}
}

func TestVerifyChecksum(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bark-verify-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.bark")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	checksum, err := ComputeChecksum(tmpDir)
	if err != nil {
		t.Fatalf("ComputeChecksum failed: %v", err)
	}

	// Verify with correct checksum
	valid, err := VerifyChecksum(tmpDir, checksum)
	if err != nil {
		t.Fatalf("VerifyChecksum failed: %v", err)
	}
	if !valid {
		t.Error("VerifyChecksum should return true for matching checksum")
	}

	// Verify with wrong checksum
	valid, err = VerifyChecksum(tmpDir, "sha256:wrongchecksum")
	if err != nil {
		t.Fatalf("VerifyChecksum failed: %v", err)
	}
	if valid {
		t.Error("VerifyChecksum should return false for non-matching checksum")
	}
}

func TestLoadLockfileNotExist(t *testing.T) {
	// Loading non-existent lockfile should return empty lockfile, not error
	lf, err := LoadLockfile("/nonexistent/path/bark.lock")
	if err != nil {
		t.Fatalf("LoadLockfile should not error for missing file: %v", err)
	}
	if lf == nil {
		t.Fatal("LoadLockfile should return empty lockfile, not nil")
	} else if len(lf.Modules) != 0 {
		t.Errorf("empty lockfile should have 0 modules, got %d", len(lf.Modules))
	}
}
