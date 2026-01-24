package manifest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitClone clones a git repository to the specified directory
func GitClone(url, dest string) error {
	cmd := exec.Command("git", "clone", "--depth=1", url, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

// GitCloneAt clones a git repository at a specific tag, branch, or commit
func GitCloneAt(url, dest string, dep *Dependency) error {
	// For tags and branches, we can use --branch with --depth=1
	if dep.Tag != "" || dep.Branch != "" {
		ref := dep.Tag
		if ref == "" {
			ref = dep.Branch
		}

		cmd := exec.Command("git", "clone", "--depth=1", "--branch", ref, url, dest)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}

		return nil
	}

	// For specific commits, we need a different approach
	if dep.Rev != "" {
		// First clone without depth limit (needed for checkout)
		cmd := exec.Command("git", "clone", url, dest)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}

		// Then checkout the specific revision
		cmd = exec.Command("git", "-C", dest, "checkout", dep.Rev)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git checkout failed: %w", err)
		}

		return nil
	}

	// Default: clone main branch
	return GitClone(url, dest)
}

// GetCacheDir returns the bark module cache directory
func GetCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".bark", "modules")
	return cacheDir, nil
}

// EnsureCacheDir creates the cache directory if it doesn't exist
func EnsureCacheDir() (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	return cacheDir, nil
}

// GetDependencyPath returns the cache path for a dependency
func GetDependencyPath(name string, dep *Dependency) (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}

	version := dep.Version()
	// Sanitize version for filesystem (replace / with _)
	version = strings.ReplaceAll(version, "/", "_")

	return filepath.Join(cacheDir, name, version), nil
}

// FetchDependency downloads a dependency if not already cached
func FetchDependency(name string, dep *Dependency, lockfile *Lockfile) (string, error) {
	// Path dependencies don't need fetching
	if dep.IsPath() {
		return dep.Path, nil
	}

	// Get cache path
	depPath, err := GetDependencyPath(name, dep)
	if err != nil {
		return "", err
	}

	// Check if already cached
	if _, err := os.Stat(depPath); err == nil {
		// Directory exists, check lockfile
		if lockfile != nil {
			locked := lockfile.GetModule(name)
			if locked != nil {
				// Verify checksum
				valid, err := VerifyChecksum(depPath, locked.Checksum)
				if err == nil && valid {
					return depPath, nil // Cache hit with valid checksum
				}
				// Checksum mismatch - re-fetch
				_ = os.RemoveAll(depPath)
			}
		} else {
			// No lockfile, trust the cache
			return depPath, nil
		}
	}

	// Ensure cache directory exists
	if _, err := EnsureCacheDir(); err != nil {
		return "", err
	}

	// Clone the repository
	fmt.Printf("Fetching %s from %s...\n", name, dep.Git)
	if err := GitCloneAt(dep.Git, depPath, dep); err != nil {
		return "", fmt.Errorf("failed to fetch %s: %w", name, err)
	}

	return depPath, nil
}

// FetchAllDependencies fetches all dependencies from a manifest
func FetchAllDependencies(m *Manifest, lockfile *Lockfile) (*Lockfile, error) {
	if lockfile == nil {
		lockfile = NewLockfile()
	}

	for name, dep := range m.Dependencies {
		depCopy := dep // Create copy for pointer
		depPath, err := FetchDependency(name, &depCopy, lockfile)
		if err != nil {
			return lockfile, fmt.Errorf("failed to fetch %s: %w", name, err)
		}

		// Update lockfile with new/verified entry
		if dep.IsGit() {
			checksum, err := ComputeChecksum(depPath)
			if err != nil {
				return lockfile, fmt.Errorf("failed to compute checksum for %s: %w", name, err)
			}

			lockfile.SetModule(LockedModule{
				Name:     name,
				Version:  dep.Version(),
				Source:   BuildSource(&depCopy),
				Checksum: checksum,
			})
		}
	}

	return lockfile, nil
}

// CleanCache removes all cached modules
func CleanCache() error {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return err
	}

	return os.RemoveAll(cacheDir)
}

// RemoveCachedDependency removes a specific cached dependency
func RemoveCachedDependency(name string) error {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return err
	}

	depDir := filepath.Join(cacheDir, name)
	return os.RemoveAll(depDir)
}
