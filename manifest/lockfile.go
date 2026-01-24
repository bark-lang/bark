package manifest

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/BurntSushi/toml"
)

// Lockfile represents a bark.lock file
type Lockfile struct {
	Modules []LockedModule `toml:"module"`
}

// LockedModule represents a locked dependency
type LockedModule struct {
	Name     string `toml:"name"`
	Version  string `toml:"version"`
	Source   string `toml:"source"`
	Checksum string `toml:"checksum"`
}

// LoadLockfile reads and parses a bark.lock file
func LoadLockfile(path string) (*Lockfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Lockfile{Modules: []LockedModule{}}, nil
		}
		return nil, fmt.Errorf("failed to read lockfile: %w", err)
	}

	return ParseLockfile(data)
}

// ParseLockfile parses bark.lock content from bytes
func ParseLockfile(data []byte) (*Lockfile, error) {
	var lf Lockfile
	if err := toml.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("failed to parse lockfile: %w", err)
	}

	if lf.Modules == nil {
		lf.Modules = []LockedModule{}
	}

	return &lf, nil
}

// GetModule returns a locked module by name, or nil if not found
func (lf *Lockfile) GetModule(name string) *LockedModule {
	for i := range lf.Modules {
		if lf.Modules[i].Name == name {
			return &lf.Modules[i]
		}
	}
	return nil
}

// HasModule returns true if the named module is in the lockfile
func (lf *Lockfile) HasModule(name string) bool {
	return lf.GetModule(name) != nil
}

// SetModule adds or updates a module in the lockfile
func (lf *Lockfile) SetModule(mod LockedModule) {
	for i := range lf.Modules {
		if lf.Modules[i].Name == mod.Name {
			lf.Modules[i] = mod
			return
		}
	}
	lf.Modules = append(lf.Modules, mod)
}

// RemoveModule removes a module from the lockfile
func (lf *Lockfile) RemoveModule(name string) {
	for i := range lf.Modules {
		if lf.Modules[i].Name == name {
			lf.Modules = append(lf.Modules[:i], lf.Modules[i+1:]...)
			return
		}
	}
}

// ToTOML returns the lockfile as a TOML string
func (lf *Lockfile) ToTOML() string {
	if len(lf.Modules) == 0 {
		return "# bark.lock - automatically generated, do not edit\n"
	}

	// Sort modules by name for deterministic output
	sorted := make([]LockedModule, len(lf.Modules))
	copy(sorted, lf.Modules)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	result := "# bark.lock - automatically generated, do not edit\n\n"
	for _, mod := range sorted {
		result += "[[module]]\n"
		result += fmt.Sprintf("name = %q\n", mod.Name)
		result += fmt.Sprintf("version = %q\n", mod.Version)
		result += fmt.Sprintf("source = %q\n", mod.Source)
		result += fmt.Sprintf("checksum = %q\n", mod.Checksum)
		result += "\n"
	}

	return result
}

// Save writes the lockfile to a file
func (lf *Lockfile) Save(path string) error {
	content := lf.ToTOML()
	return os.WriteFile(path, []byte(content), 0644)
}

// FindLockfile searches for bark.lock starting from the given directory
func FindLockfile(startDir string) (string, error) {
	dir := startDir
	for {
		lockPath := filepath.Join(dir, "bark.lock")
		if _, err := os.Stat(lockPath); err == nil {
			return lockPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("bark.lock not found in %s or any parent directory", startDir)
		}
		dir = parent
	}
}

// BuildSource constructs the source string for a git dependency
func BuildSource(dep *Dependency) string {
	if dep.IsPath() {
		return "path+" + dep.Path
	}

	source := "git+" + dep.Git
	if dep.Tag != "" {
		source += "?tag=" + dep.Tag
	} else if dep.Branch != "" {
		source += "?branch=" + dep.Branch
	} else if dep.Rev != "" {
		source += "?rev=" + dep.Rev
	}
	return source
}

// ComputeChecksum calculates a SHA256 checksum for a directory
func ComputeChecksum(dir string) (string, error) {
	h := sha256.New()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and hidden files
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if info.Name()[0] == '.' {
			return nil
		}

		// Only hash .bark files
		if filepath.Ext(path) != ".bark" {
			return nil
		}

		// Hash file content
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Include relative path in hash for structure awareness
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		h.Write([]byte(relPath))
		h.Write(data)

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to compute checksum: %w", err)
	}

	return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil
}

// VerifyChecksum verifies that a directory matches the expected checksum
func VerifyChecksum(dir, expected string) (bool, error) {
	actual, err := ComputeChecksum(dir)
	if err != nil {
		return false, err
	}
	return actual == expected, nil
}

// NewLockfile creates an empty lockfile
func NewLockfile() *Lockfile {
	return &Lockfile{Modules: []LockedModule{}}
}
