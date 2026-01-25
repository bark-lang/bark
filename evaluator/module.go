package evaluator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/bark-lang/bark/ast"
	"gitlab.com/bark-lang/bark/lexer"
	"gitlab.com/bark-lang/bark/manifest"
	"gitlab.com/bark-lang/bark/object"
	"gitlab.com/bark-lang/bark/parser"
)

// Module represents a loaded bark module
type Module struct {
	Name        string              // Module name from 'module' declaration
	Path        string              // File path of the module
	Environment *object.Environment // Module's environment
	PublicFuncs map[string]bool     // Set of public function names
	AST         *ast.Program        // Parsed AST
	Evaluated   bool                // Whether module has been evaluated
}

// ModuleRegistry manages all loaded modules
type ModuleRegistry struct {
	modules       map[string]*Module // Key: module path
	importAliases map[string]string  // Key: alias, Value: module name or module path
	currentModule string             // Current module being evaluated (for tracking context)
	currentFile   string             // Current file path (for resolving relative imports)
	loadingStack  []string           // Stack for circular dependency detection
	manifest      *manifest.Manifest // Project manifest (bark.toml)
	lockfile      *manifest.Lockfile // Lockfile (bark.lock)
	projectRoot   string             // Root directory of the project
}

// NewModuleRegistry creates a new module registry
func NewModuleRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		modules:       make(map[string]*Module),
		importAliases: make(map[string]string),
		loadingStack:  []string{},
	}
}

// SetManifest sets the project manifest
func (mr *ModuleRegistry) SetManifest(m *manifest.Manifest) {
	mr.manifest = m
}

// GetManifest returns the project manifest
func (mr *ModuleRegistry) GetManifest() *manifest.Manifest {
	return mr.manifest
}

// SetLockfile sets the project lockfile
func (mr *ModuleRegistry) SetLockfile(lf *manifest.Lockfile) {
	mr.lockfile = lf
}

// GetLockfile returns the project lockfile
func (mr *ModuleRegistry) GetLockfile() *manifest.Lockfile {
	return mr.lockfile
}

// SetProjectRoot sets the project root directory
func (mr *ModuleRegistry) SetProjectRoot(root string) {
	mr.projectRoot = root
}

// GetProjectRoot returns the project root directory
func (mr *ModuleRegistry) GetProjectRoot() string {
	return mr.projectRoot
}

// IsLoading checks if a module is currently being loaded (for circular dependency detection)
func (mr *ModuleRegistry) IsLoading(path string) bool {
	for _, p := range mr.loadingStack {
		if p == path {
			return true
		}
	}
	return false
}

// PushLoading adds a module to the loading stack
func (mr *ModuleRegistry) PushLoading(path string) {
	mr.loadingStack = append(mr.loadingStack, path)
}

// PopLoading removes the most recent module from the loading stack
func (mr *ModuleRegistry) PopLoading() {
	if len(mr.loadingStack) > 0 {
		mr.loadingStack = mr.loadingStack[:len(mr.loadingStack)-1]
	}
}

// GetLoadingPath returns the circular dependency path
func (mr *ModuleRegistry) GetLoadingPath() string {
	return strings.Join(mr.loadingStack, " -> ")
}

// GetModule retrieves a module by path
func (mr *ModuleRegistry) GetModule(path string) (*Module, bool) {
	mod, ok := mr.modules[path]
	return mod, ok
}

// RegisterModule adds a module to the registry
func (mr *ModuleRegistry) RegisterModule(path string, mod *Module) {
	mr.modules[path] = mod
}

// SetAlias registers an import alias
func (mr *ModuleRegistry) SetAlias(alias, moduleName string) {
	mr.importAliases[alias] = moduleName
}

// ResolveAlias resolves an alias to a module name
func (mr *ModuleRegistry) ResolveAlias(name string) string {
	if alias, ok := mr.importAliases[name]; ok {
		return alias
	}
	return name
}

// SetCurrentModule sets the current module being evaluated
func (mr *ModuleRegistry) SetCurrentModule(name string) {
	mr.currentModule = name
}

// GetCurrentModule returns the current module name
func (mr *ModuleRegistry) GetCurrentModule() string {
	return mr.currentModule
}

// SetCurrentFile sets the current file path being evaluated
func (mr *ModuleRegistry) SetCurrentFile(path string) {
	mr.currentFile = path
}

// GetCurrentFile returns the current file path
func (mr *ModuleRegistry) GetCurrentFile() string {
	return mr.currentFile
}

// ResolvePath resolves a module import path to an absolute file path
// This is the legacy function for backward compatibility
func ResolvePath(importPath, currentFilePath string) (string, error) {
	return ResolvePathWithRegistry(importPath, currentFilePath, nil)
}

// ResolvePathWithRegistry resolves a module import path using the manifest if available
func ResolvePathWithRegistry(importPath, currentFilePath string, registry *ModuleRegistry) (string, error) {
	// Check if this is a dependency from bark.toml
	if registry != nil && registry.manifest != nil {
		// Extract the base module name (first part before /)
		baseName, subPath, _ := strings.Cut(importPath, "/")

		if dep := registry.manifest.GetDependency(baseName); dep != nil {
			return resolveDependencyPath(baseName, subPath, dep, registry)
		}
	}

	// Fall back to local file resolution
	return resolveLocalPath(importPath, currentFilePath)
}

// resolveDependencyPath resolves a path for a manifest dependency
func resolveDependencyPath(name, subPath string, dep *manifest.Dependency, registry *ModuleRegistry) (string, error) {
	var basePath string

	if dep.IsPath() {
		// Path dependency - resolve relative to project root
		if registry.projectRoot != "" {
			basePath = filepath.Join(registry.projectRoot, dep.Path)
		} else {
			basePath = dep.Path
		}
	} else if dep.IsGit() {
		// Git dependency - fetch if needed and get cache path
		depPath, err := manifest.FetchDependency(name, dep, registry.lockfile)
		if err != nil {
			return "", fmt.Errorf("failed to fetch dependency %s: %w", name, err)
		}
		basePath = depPath
	} else {
		return "", fmt.Errorf("invalid dependency %s: must specify 'git' or 'path'", name)
	}

	// If there's a subpath, append it
	if subPath != "" {
		basePath = filepath.Join(basePath, subPath)
	}

	// Try with .bark extension
	if !strings.HasSuffix(basePath, ".bark") {
		withExt := basePath + ".bark"
		if _, err := os.Stat(withExt); err == nil {
			return withExt, nil
		}

		// Try as directory with module.bark
		modulePath := filepath.Join(basePath, "module.bark")
		if _, err := os.Stat(modulePath); err == nil {
			return modulePath, nil
		}

		// If no subpath, look for main module file
		if subPath == "" {
			// Try <name>.bark in the directory
			mainFile := filepath.Join(basePath, name+".bark")
			if _, err := os.Stat(mainFile); err == nil {
				return mainFile, nil
			}
		}
	}

	// Check if the path exists as-is
	if _, err := os.Stat(basePath); err == nil {
		return basePath, nil
	}

	return "", fmt.Errorf("module not found: %s (looked in %s)", name, basePath)
}

// resolveLocalPath resolves a local file path
func resolveLocalPath(importPath, currentFilePath string) (string, error) {
	// Get the directory of the current file
	baseDir := filepath.Dir(currentFilePath)
	if currentFilePath == "" {
		// If no current file, use current working directory
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		baseDir = wd
	}

	// Resolve relative to base directory
	absPath := filepath.Join(baseDir, importPath)

	// Try with .bark extension if not present
	if !strings.HasSuffix(absPath, ".bark") {
		absPath += ".bark"
	}

	// Check if file exists
	if _, err := os.Stat(absPath); err != nil {
		// Try as directory with module.bark
		dirPath := strings.TrimSuffix(absPath, ".bark")
		modulePath := filepath.Join(dirPath, "module.bark")
		if _, err := os.Stat(modulePath); err == nil {
			return modulePath, nil
		}
		return "", fmt.Errorf("module not found: %s", importPath)
	}

	return absPath, nil
}

// LoadManifestForFile attempts to find and load a bark.toml for the given file
func LoadManifestForFile(filePath string) (*manifest.Manifest, *manifest.Lockfile, string, error) {
	dir := filepath.Dir(filePath)
	if filePath == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, nil, "", err
		}
	}

	// Try to find bark.toml
	manifestPath, err := manifest.FindManifest(dir)
	if err != nil {
		// No manifest found - that's OK, not all projects need one
		return nil, nil, "", nil
	}

	// Load manifest
	m, err := manifest.Load(manifestPath)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to load manifest: %w", err)
	}

	projectRoot := filepath.Dir(manifestPath)

	// Try to load lockfile
	lockPath := filepath.Join(projectRoot, "bark.lock")
	lf, _ := manifest.LoadLockfile(lockPath) // Ignore error - lockfile is optional

	return m, lf, projectRoot, nil
}

// LoadModule loads and parses a module file
func LoadModule(path string, registry *ModuleRegistry) (*Module, error) {
	// Check for circular dependencies FIRST (even if module exists but is being evaluated)
	if registry.IsLoading(path) {
		return nil, fmt.Errorf("circular dependency detected: %s -> %s",
			registry.GetLoadingPath(), path)
	}

	// Check if already loaded AND evaluated
	if mod, ok := registry.GetModule(path); ok && mod.Evaluated {
		return mod, nil
	}

	// Read the file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read module %s: %v", path, err)
	}

	// Parse the file
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors in %s: %v", path, p.Errors())
	}

	// Extract module name from module statement (if present)
	moduleName := "_" // Default module name
	for _, stmt := range program.Statements {
		if modStmt, ok := stmt.(*ast.ModuleStatement); ok {
			moduleName = modStmt.Name.Value
			break
		}
	}

	// Create module
	mod := &Module{
		Name:        moduleName,
		Path:        path,
		Environment: object.NewEnvironment(),
		PublicFuncs: make(map[string]bool),
		AST:         program,
		Evaluated:   false,
	}

	// Register the module
	registry.RegisterModule(path, mod)

	return mod, nil
}

// EvaluateModule evaluates a module's AST in its environment
func EvaluateModule(mod *Module, registry *ModuleRegistry) error {
	if mod.Evaluated {
		return nil // Already evaluated
	}

	// Mark module as being loaded (for circular dependency detection)
	registry.PushLoading(mod.Path)
	defer registry.PopLoading()

	// Set current module context
	oldModule := registry.GetCurrentModule()
	registry.SetCurrentModule(mod.Name)
	defer registry.SetCurrentModule(oldModule)

	// Set current file for import resolution
	oldFile := registry.GetCurrentFile()
	registry.SetCurrentFile(mod.Path)
	defer registry.SetCurrentFile(oldFile)

	// Evaluate each statement in the module
	for _, stmt := range mod.AST.Statements {
		result := Eval(stmt, mod.Environment)
		if isError(result) {
			return fmt.Errorf("error evaluating module %s: %s", mod.Name, result.Inspect())
		}

		// Track public functions
		if fnStmt, ok := stmt.(*ast.FunctionStatement); ok {
			if fnStmt.Public {
				mod.PublicFuncs[fnStmt.Name.Value] = true
			}
		}
	}

	mod.Evaluated = true
	return nil
}

// IsPublicFunction checks if a function is public in a module
func (mod *Module) IsPublicFunction(name string) bool {
	return mod.PublicFuncs[name]
}

// GetFunction retrieves a function from the module's environment
func (mod *Module) GetFunction(name string) (object.Object, bool) {
	obj, ok := mod.Environment.Get(name)
	return obj, ok
}
