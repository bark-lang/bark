package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/bark-lang/bark/bytecode"
	"gitlab.com/bark-lang/bark/evaluator"
	"gitlab.com/bark-lang/bark/lexer"
	"gitlab.com/bark-lang/bark/manifest"
	"gitlab.com/bark-lang/bark/object"
	"gitlab.com/bark-lang/bark/parser"
)

// useBytecode controls whether to use the bytecode VM or tree-walking interpreter
var useBytecode bool
var showDisasm bool

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Parse flags manually for flexibility
	args := os.Args[1:]
	var fileArg string
	var codeArg string
	isEval := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--bytecode", "-b":
			useBytecode = true
		case "--disasm", "-d":
			showDisasm = true
			useBytecode = true // disasm implies bytecode
		case "-e", "--eval":
			isEval = true
			if i+1 < len(args) {
				i++
				codeArg = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "Error: -e requires code argument\n")
				os.Exit(1)
			}
		case "test":
			// Collect remaining args as test paths
			testArgs := args[i+1:]
			cmdTest(testArgs)
			return
		case "check":
			// Collect remaining args as paths to check
			checkArgs := args[i+1:]
			cmdCheck(checkArgs)
			return
		case "init":
			cmdInit()
			return
		case "get":
			cmdGet()
			return
		case "update":
			cmdUpdate()
			return
		case "help", "-h", "--help":
			printUsage()
			return
		case "version", "-v", "--version":
			fmt.Println("bark version 0.1.0")
			return
		default:
			if fileArg == "" && !isEval {
				fileArg = args[i]
			}
		}
	}

	if isEval && codeArg != "" {
		runCode(codeArg)
	} else if fileArg != "" {
		runFile(fileArg)
	} else {
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Bark Programming Language")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  bark <file.bark>    Run a Bark program")
	fmt.Println("  bark -e <code>      Execute Bark code directly")
	fmt.Println("  bark check          Check all .bark files in current directory parse correctly")
	fmt.Println("  bark check <path>   Check .bark files in specified path recursively")
	fmt.Println("  bark test           Run all tests in tests/ directory")
	fmt.Println("  bark test <path>    Run a specific test file or directory")
	fmt.Println("  bark init           Create a new bark.toml")
	fmt.Println("  bark get            Fetch dependencies")
	fmt.Println("  bark update         Update dependencies")
	fmt.Println("  bark help           Show this help message")
	fmt.Println("  bark version        Show version information")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -b, --bytecode      Use bytecode VM instead of tree-walker")
	fmt.Println("  -d, --disasm        Show bytecode disassembly (implies --bytecode)")
}

func runFile(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	p.SetFile(filename)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			fmt.Fprint(os.Stderr, err.FormatError())
		}
		os.Exit(1)
	}

	if useBytecode {
		// Bytecode VM path
		compileResult := bytecode.Compile(program)
		if len(compileResult.Errors) > 0 {
			fmt.Fprintf(os.Stderr, "Compilation errors:\n")
			for _, msg := range compileResult.Errors {
				fmt.Fprintf(os.Stderr, "  %s\n", msg)
			}
			os.Exit(1)
		}

		if showDisasm {
			fmt.Println(bytecode.Disassemble(compileResult.Function.Chunk, filename))
		}

		vm := bytecode.NewVM()
		result, err := vm.Run(compileResult.Function)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
			os.Exit(1)
		}

		if result != nil && result.Type() != object.NULL_OBJ {
			_, _ = io.WriteString(os.Stdout, result.Inspect())
			_, _ = io.WriteString(os.Stdout, "\n")
		}
		return
	}

	// Tree-walking interpreter path
	// Set current file for module resolution before evaluating
	evaluator.SetCurrentFile(filename)

	// Set source context for error reporting
	evaluator.SetSourceContext(filename, string(content))

	// Try to load manifest for the project
	m, lf, projectRoot, err := evaluator.LoadManifestForFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	// Set up module registry with manifest
	if m != nil {
		registry := evaluator.GetModuleRegistry()
		registry.SetManifest(m)
		registry.SetLockfile(lf)
		registry.SetProjectRoot(projectRoot)
	}

	env := object.NewEnvironment()
	result := evaluator.Eval(program, env)

	if result != nil {
		// Only programming errors are fatal - Bark error values can be returned
		if result.Type() == object.ERROR_OBJ {
			if errObj, ok := result.(*object.Error); ok && errObj.IsProgrammingError {
				fmt.Fprint(os.Stderr, errObj.FormatError())
				os.Exit(1)
			}
		}
		// Only print non-null results
		if result.Type() != object.NULL_OBJ {
			_, _ = io.WriteString(os.Stdout, result.Inspect())
			_, _ = io.WriteString(os.Stdout, "\n")
		}
	}
}

func runCode(code string) {
	l := lexer.New(code)
	p := parser.New(l)
	p.SetFile("<stdin>")
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			fmt.Fprint(os.Stderr, err.FormatError())
		}
		os.Exit(1)
	}

	if useBytecode {
		// Bytecode VM path
		compileResult := bytecode.Compile(program)
		if len(compileResult.Errors) > 0 {
			fmt.Fprintf(os.Stderr, "Compilation errors:\n")
			for _, msg := range compileResult.Errors {
				fmt.Fprintf(os.Stderr, "  %s\n", msg)
			}
			os.Exit(1)
		}

		if showDisasm {
			fmt.Println(bytecode.Disassemble(compileResult.Function.Chunk, "<eval>"))
		}

		vm := bytecode.NewVM()
		result, err := vm.Run(compileResult.Function)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
			os.Exit(1)
		}

		if result != nil && result.Type() != object.NULL_OBJ {
			_, _ = io.WriteString(os.Stdout, result.Inspect())
			_, _ = io.WriteString(os.Stdout, "\n")
		}
		return
	}

	// Tree-walking interpreter path
	// Set source context for error reporting
	evaluator.SetSourceContext("<eval>", code)

	env := object.NewEnvironment()
	result := evaluator.Eval(program, env)

	if result != nil {
		// Only programming errors are fatal - Bark error values can be returned
		if result.Type() == object.ERROR_OBJ {
			if errObj, ok := result.(*object.Error); ok && errObj.IsProgrammingError {
				fmt.Fprint(os.Stderr, errObj.FormatError())
				os.Exit(1)
			}
		}
		// Only print non-null results
		if result.Type() != object.NULL_OBJ {
			_, _ = io.WriteString(os.Stdout, result.Inspect())
			_, _ = io.WriteString(os.Stdout, "\n")
		}
	}
}

func cmdInit() {
	// Check if bark.toml already exists
	if _, err := os.Stat("bark.toml"); err == nil {
		fmt.Fprintf(os.Stderr, "Error: bark.toml already exists\n")
		os.Exit(1)
	}

	// Get directory name for module name
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	moduleName := filepath.Base(wd)

	// Create default manifest
	m := manifest.DefaultManifest(moduleName)

	// Save to bark.toml
	if err := m.Save("bark.toml"); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating bark.toml: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created bark.toml for module %q\n", moduleName)
}

func cmdGet() {
	// Find and load manifest
	manifestPath, err := manifest.FindManifest(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: bark.toml not found. Run 'bark init' first.\n")
		os.Exit(1)
	}

	m, err := manifest.Load(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading bark.toml: %v\n", err)
		os.Exit(1)
	}

	if err := m.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error in bark.toml: %v\n", err)
		os.Exit(1)
	}

	projectRoot := filepath.Dir(manifestPath)
	lockPath := filepath.Join(projectRoot, "bark.lock")

	// Load existing lockfile
	lf, _ := manifest.LoadLockfile(lockPath)

	// Fetch all dependencies
	fmt.Println("Fetching dependencies...")
	lf, err = manifest.FetchAllDependencies(m, lf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching dependencies: %v\n", err)
		os.Exit(1)
	}

	// Save updated lockfile
	if err := lf.Save(lockPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving bark.lock: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fetched %d dependencies\n", len(m.Dependencies))
}

func cmdUpdate() {
	// Find and load manifest
	manifestPath, err := manifest.FindManifest(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: bark.toml not found. Run 'bark init' first.\n")
		os.Exit(1)
	}

	m, err := manifest.Load(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading bark.toml: %v\n", err)
		os.Exit(1)
	}

	if err := m.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error in bark.toml: %v\n", err)
		os.Exit(1)
	}

	projectRoot := filepath.Dir(manifestPath)
	lockPath := filepath.Join(projectRoot, "bark.lock")

	// Clear cache for all dependencies to force re-fetch
	fmt.Println("Updating dependencies...")
	for name := range m.Dependencies {
		_ = manifest.RemoveCachedDependency(name)
	}

	// Fetch all dependencies fresh
	lf, err := manifest.FetchAllDependencies(m, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating dependencies: %v\n", err)
		os.Exit(1)
	}

	// Save new lockfile
	if err := lf.Save(lockPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving bark.lock: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Updated %d dependencies\n", len(m.Dependencies))
}

func cmdTest(args []string) {
	var testPaths []string

	if len(args) == 0 {
		// Default: run all tests in tests/ directory
		if _, err := os.Stat("tests"); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: tests/ directory not found\n")
			fmt.Fprintf(os.Stderr, "Create a tests/ directory with .bark test files, or specify a path:\n")
			fmt.Fprintf(os.Stderr, "  bark test tests/<path>\n")
			os.Exit(1)
		}
		testPaths = []string{"tests"}
	} else {
		// Validate that all paths are within tests/ directory
		for _, p := range args {
			if !isInTestsDir(p) {
				fmt.Fprintf(os.Stderr, "Error: %s is not in tests/ directory\n", p)
				fmt.Fprintf(os.Stderr, "Test files must be in the tests/ directory\n")
				os.Exit(1)
			}
		}
		testPaths = args
	}

	var allFiles []string
	for _, path := range testPaths {
		files, err := collectTestFiles(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		allFiles = append(allFiles, files...)
	}

	if len(allFiles) == 0 {
		fmt.Println("No test files found")
		return
	}

	// Enable test mode for assert() and assert_error() builtins
	evaluator.SetTestMode(true)
	defer evaluator.SetTestMode(false)

	passed := 0
	failed := 0

	for _, file := range allFiles {
		ok := runTestFile(file)
		if ok {
			fmt.Printf("PASS %s\n", file)
			passed++
		} else {
			fmt.Printf("FAIL %s\n", file)
			failed++
		}
	}

	fmt.Println()
	fmt.Printf("Results: %d passed, %d failed\n", passed, failed)

	if failed > 0 {
		os.Exit(1)
	}
}

// isInTestsDir checks if a path is within the tests/ directory
func isInTestsDir(path string) bool {
	// Clean the path to handle . and ..
	cleanPath := filepath.Clean(path)

	// Check if path starts with "tests/" or is "tests"
	if cleanPath == "tests" {
		return true
	}
	if strings.HasPrefix(cleanPath, "tests/") || strings.HasPrefix(cleanPath, "tests\\") {
		return true
	}

	return false
}

// collectTestFiles finds all .bark files in the given path
func collectTestFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot access %s: %w", path, err)
	}

	if !info.IsDir() {
		// Single file
		if filepath.Ext(path) != ".bark" {
			return nil, fmt.Errorf("%s is not a .bark file", path)
		}
		return []string{path}, nil
	}

	// Directory: walk and collect all .bark files
	var files []string
	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(p) == ".bark" {
			files = append(files, p)
		}
		return nil
	})

	return files, err
}

// runTestFile runs a single test file and returns true if it passed
func runTestFile(filename string) bool {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error reading file: %v\n", err)
		return false
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	p.SetFile(filename)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			fmt.Fprint(os.Stderr, err.FormatError())
		}
		return false
	}

	// Set up evaluator context
	evaluator.SetCurrentFile(filename)
	evaluator.SetSourceContext(filename, string(content))

	// Try to load manifest for the project
	m, lf, projectRoot, err := evaluator.LoadManifestForFile(filename)
	if err == nil && m != nil {
		registry := evaluator.GetModuleRegistry()
		registry.SetManifest(m)
		registry.SetLockfile(lf)
		registry.SetProjectRoot(projectRoot)
	}

	// Capture stderr to detect assertion failures
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	env := object.NewEnvironment()
	result := evaluator.Eval(program, env)

	// Restore stderr and read captured output
	_ = w.Close()
	os.Stderr = oldStderr

	var stderrBuf bytes.Buffer
	_, _ = stderrBuf.ReadFrom(r)
	stderrOutput := stderrBuf.String()

	// Print any stderr output (assertion failures, etc.)
	if stderrOutput != "" {
		fmt.Fprint(os.Stderr, stderrOutput)
	}

	// Check for assertion failures in stderr
	hasAssertionFailure := strings.Contains(stderrOutput, "ASSERTION FAILED")

	// Check for programming errors
	if result != nil && result.Type() == object.ERROR_OBJ {
		if errObj, ok := result.(*object.Error); ok && errObj.IsProgrammingError {
			fmt.Fprint(os.Stderr, errObj.FormatError())
			return false
		}
	}

	return !hasAssertionFailure
}

func cmdCheck(args []string) {
	var checkPaths []string

	if len(args) == 0 {
		// Default: check current directory
		checkPaths = []string{"."}
	} else {
		checkPaths = args
	}

	var allFiles []string
	for _, path := range checkPaths {
		files, err := collectBarkFiles(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		allFiles = append(allFiles, files...)
	}

	if len(allFiles) == 0 {
		fmt.Println("No .bark files found")
		return
	}

	passed := 0
	failed := 0
	var failedFiles []string

	for _, file := range allFiles {
		ok := checkFile(file)
		if ok {
			passed++
		} else {
			failed++
			failedFiles = append(failedFiles, file)
		}
	}

	fmt.Println()
	if failed == 0 {
		fmt.Printf("All %d files parsed successfully\n", passed)
	} else {
		fmt.Printf("Results: %d passed, %d failed\n", passed, failed)
		os.Exit(1)
	}
}

// collectBarkFiles finds all .bark files in the given path recursively
func collectBarkFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot access %s: %w", path, err)
	}

	if !info.IsDir() {
		// Single file
		if filepath.Ext(path) != ".bark" {
			return nil, fmt.Errorf("%s is not a .bark file", path)
		}
		return []string{path}, nil
	}

	// Directory: walk and collect all .bark files
	var files []string
	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(p) == ".bark" {
			files = append(files, p)
		}
		return nil
	})

	return files, err
}

// checkFile parses a single file and returns true if it parsed successfully
func checkFile(filename string) bool {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filename, err)
		return false
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	p.SetFile(filename)
	_ = p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "FAIL %s\n", filename)
		for _, err := range p.Errors() {
			fmt.Fprint(os.Stderr, err.FormatError())
		}
		return false
	}

	return true
}
