// Package manifest handles parsing and managing bark.toml manifest files.
package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Manifest represents a bark.toml file
type Manifest struct {
	Module       ModuleInfo            `toml:"module"`
	Dependencies map[string]Dependency `toml:"dependencies"`
}

// ModuleInfo contains module metadata
type ModuleInfo struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
}

// Dependency represents a single dependency in bark.toml
type Dependency struct {
	// Git dependency fields
	Git    string `toml:"git"`
	Tag    string `toml:"tag"`
	Branch string `toml:"branch"`
	Rev    string `toml:"rev"`

	// Path dependency field
	Path string `toml:"path"`
}

// IsGit returns true if this is a git dependency
func (d *Dependency) IsGit() bool {
	return d.Git != ""
}

// IsPath returns true if this is a path dependency
func (d *Dependency) IsPath() bool {
	return d.Path != ""
}

// Version returns the version identifier (tag, branch, or rev)
func (d *Dependency) Version() string {
	if d.Tag != "" {
		return d.Tag
	}
	if d.Branch != "" {
		return d.Branch
	}
	if d.Rev != "" {
		return d.Rev
	}
	return "main" // default to main branch
}

// VersionType returns the type of version specifier used
func (d *Dependency) VersionType() string {
	if d.Tag != "" {
		return "tag"
	}
	if d.Branch != "" {
		return "branch"
	}
	if d.Rev != "" {
		return "rev"
	}
	return "branch" // default
}

// Load reads and parses a bark.toml file from the given path
func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	return Parse(data)
}

// Parse parses bark.toml content from bytes
func Parse(data []byte) (*Manifest, error) {
	var m Manifest
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Initialize empty dependencies map if nil
	if m.Dependencies == nil {
		m.Dependencies = make(map[string]Dependency)
	}

	return &m, nil
}

// FindManifest searches for bark.toml starting from the given directory
// and walking up to parent directories until found or root is reached
func FindManifest(startDir string) (string, error) {
	dir := startDir
	for {
		manifestPath := filepath.Join(dir, "bark.toml")
		if _, err := os.Stat(manifestPath); err == nil {
			return manifestPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding manifest
			return "", fmt.Errorf("bark.toml not found in %s or any parent directory", startDir)
		}
		dir = parent
	}
}

// GetDependency returns a dependency by name, or nil if not found
func (m *Manifest) GetDependency(name string) *Dependency {
	if dep, ok := m.Dependencies[name]; ok {
		return &dep
	}
	return nil
}

// HasDependency returns true if the named dependency exists
func (m *Manifest) HasDependency(name string) bool {
	_, ok := m.Dependencies[name]
	return ok
}

// Validate checks the manifest for errors
func (m *Manifest) Validate() error {
	if m.Module.Name == "" {
		return fmt.Errorf("module name is required in [module] section")
	}

	for name, dep := range m.Dependencies {
		if !dep.IsGit() && !dep.IsPath() {
			return fmt.Errorf("dependency %q must specify either 'git' or 'path'", name)
		}
		if dep.IsGit() && dep.IsPath() {
			return fmt.Errorf("dependency %q cannot specify both 'git' and 'path'", name)
		}
	}

	return nil
}

// DefaultManifest returns a new manifest with default values
func DefaultManifest(name string) *Manifest {
	return &Manifest{
		Module: ModuleInfo{
			Name:    name,
			Version: "0.1.0",
		},
		Dependencies: make(map[string]Dependency),
	}
}

// ToTOML returns the manifest as a TOML string
func (m *Manifest) ToTOML() string {
	var result string

	result += "[module]\n"
	result += fmt.Sprintf("name = %q\n", m.Module.Name)
	result += fmt.Sprintf("version = %q\n", m.Module.Version)

	if len(m.Dependencies) > 0 {
		result += "\n[dependencies]\n"
		for name, dep := range m.Dependencies {
			if dep.IsPath() {
				result += fmt.Sprintf("%s = { path = %q }\n", name, dep.Path)
			} else if dep.IsGit() {
				result += fmt.Sprintf("%s = { git = %q", name, dep.Git)
				if dep.Tag != "" {
					result += fmt.Sprintf(", tag = %q", dep.Tag)
				} else if dep.Branch != "" {
					result += fmt.Sprintf(", branch = %q", dep.Branch)
				} else if dep.Rev != "" {
					result += fmt.Sprintf(", rev = %q", dep.Rev)
				}
				result += " }\n"
			}
		}
	}

	return result
}

// Save writes the manifest to a file
func (m *Manifest) Save(path string) error {
	content := m.ToTOML()
	return os.WriteFile(path, []byte(content), 0644)
}
