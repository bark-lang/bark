package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantVer  string
		wantDeps int
		wantErr  bool
	}{
		{
			name: "basic manifest",
			input: `
[module]
name = "myapp"
version = "0.1.0"
`,
			wantName: "myapp",
			wantVer:  "0.1.0",
			wantDeps: 0,
			wantErr:  false,
		},
		{
			name: "manifest with git dependency",
			input: `
[module]
name = "myapp"
version = "1.0.0"

[dependencies]
json = { git = "https://gitlab.com/bark-lang/modules/json", tag = "v1.0.0" }
`,
			wantName: "myapp",
			wantVer:  "1.0.0",
			wantDeps: 1,
			wantErr:  false,
		},
		{
			name: "manifest with path dependency",
			input: `
[module]
name = "myapp"
version = "0.1.0"

[dependencies]
mylib = { path = "./lib/mylib" }
`,
			wantName: "myapp",
			wantVer:  "0.1.0",
			wantDeps: 1,
			wantErr:  false,
		},
		{
			name: "manifest with multiple dependencies",
			input: `
[module]
name = "myapp"
version = "0.1.0"

[dependencies]
json = { git = "https://gitlab.com/bark-lang/modules/json", tag = "v1.0.0" }
logger = { git = "https://gitlab.com/user/bark-logger", branch = "main" }
mylib = { path = "./lib/mylib" }
`,
			wantName: "myapp",
			wantVer:  "0.1.0",
			wantDeps: 3,
			wantErr:  false,
		},
		{
			name:    "invalid toml",
			input:   "this is not valid toml [[[",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := Parse([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if m.Module.Name != tt.wantName {
				t.Errorf("name = %q, want %q", m.Module.Name, tt.wantName)
			}
			if m.Module.Version != tt.wantVer {
				t.Errorf("version = %q, want %q", m.Module.Version, tt.wantVer)
			}
			if len(m.Dependencies) != tt.wantDeps {
				t.Errorf("dependencies count = %d, want %d", len(m.Dependencies), tt.wantDeps)
			}
		})
	}
}

func TestDependencyTypes(t *testing.T) {
	input := `
[module]
name = "test"
version = "0.1.0"

[dependencies]
gitdep = { git = "https://example.com/repo", tag = "v1.0.0" }
pathdep = { path = "./local" }
branchdep = { git = "https://example.com/repo2", branch = "develop" }
revdep = { git = "https://example.com/repo3", rev = "abc123" }
`
	m, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	tests := []struct {
		name        string
		isGit       bool
		isPath      bool
		version     string
		versionType string
	}{
		{"gitdep", true, false, "v1.0.0", "tag"},
		{"pathdep", false, true, "main", "branch"},
		{"branchdep", true, false, "develop", "branch"},
		{"revdep", true, false, "abc123", "rev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := m.GetDependency(tt.name)
			if dep == nil {
				t.Fatalf("dependency %q not found", tt.name)
			}

			if dep.IsGit() != tt.isGit {
				t.Errorf("IsGit() = %v, want %v", dep.IsGit(), tt.isGit)
			}
			if dep.IsPath() != tt.isPath {
				t.Errorf("IsPath() = %v, want %v", dep.IsPath(), tt.isPath)
			}
			if dep.Version() != tt.version {
				t.Errorf("Version() = %q, want %q", dep.Version(), tt.version)
			}
			if dep.VersionType() != tt.versionType {
				t.Errorf("VersionType() = %q, want %q", dep.VersionType(), tt.versionType)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid manifest",
			input: `
[module]
name = "myapp"
version = "0.1.0"

[dependencies]
json = { git = "https://example.com/json", tag = "v1.0.0" }
`,
			wantErr: false,
		},
		{
			name: "missing module name",
			input: `
[module]
version = "0.1.0"
`,
			wantErr: true,
		},
		{
			name: "dependency without git or path",
			input: `
[module]
name = "myapp"
version = "0.1.0"

[dependencies]
broken = { tag = "v1.0.0" }
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := Parse([]byte(tt.input))
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			err = m.Validate()
			if tt.wantErr && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestFindManifest(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "bark-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create nested directories
	subDir := filepath.Join(tmpDir, "src", "lib")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirs: %v", err)
	}

	// Create bark.toml at root
	manifestPath := filepath.Join(tmpDir, "bark.toml")
	if err := os.WriteFile(manifestPath, []byte("[module]\nname = \"test\"\nversion = \"0.1.0\"\n"), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	// Test finding manifest from subdirectory
	found, err := FindManifest(subDir)
	if err != nil {
		t.Fatalf("FindManifest failed: %v", err)
	}
	if found != manifestPath {
		t.Errorf("found = %q, want %q", found, manifestPath)
	}

	// Test not finding manifest
	emptyDir, err := os.MkdirTemp("", "bark-empty-*")
	if err != nil {
		t.Fatalf("failed to create empty temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(emptyDir) }()

	_, err = FindManifest(emptyDir)
	if err == nil {
		t.Error("expected error for missing manifest, got nil")
	}
}

func TestDefaultManifest(t *testing.T) {
	m := DefaultManifest("myproject")

	if m.Module.Name != "myproject" {
		t.Errorf("name = %q, want %q", m.Module.Name, "myproject")
	}
	if m.Module.Version != "0.1.0" {
		t.Errorf("version = %q, want %q", m.Module.Version, "0.1.0")
	}
	if m.Dependencies == nil {
		t.Error("dependencies should not be nil")
	}
}

func TestToTOML(t *testing.T) {
	m := &Manifest{
		Module: ModuleInfo{
			Name:    "myapp",
			Version: "1.0.0",
		},
		Dependencies: map[string]Dependency{
			"json": {Git: "https://example.com/json", Tag: "v1.0.0"},
		},
	}

	toml := m.ToTOML()

	if toml == "" {
		t.Error("ToTOML returned empty string")
	}

	// Parse it back to verify it's valid
	parsed, err := Parse([]byte(toml))
	if err != nil {
		t.Fatalf("generated TOML is not valid: %v", err)
	}

	if parsed.Module.Name != m.Module.Name {
		t.Errorf("round-trip name = %q, want %q", parsed.Module.Name, m.Module.Name)
	}
}

func TestHasDependency(t *testing.T) {
	m := &Manifest{
		Module: ModuleInfo{Name: "test", Version: "0.1.0"},
		Dependencies: map[string]Dependency{
			"json": {Git: "https://example.com/json", Tag: "v1.0.0"},
		},
	}

	if !m.HasDependency("json") {
		t.Error("HasDependency(\"json\") should return true")
	}
	if m.HasDependency("notexist") {
		t.Error("HasDependency(\"notexist\") should return false")
	}
}
