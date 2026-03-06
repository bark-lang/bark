package builtins

import (
	"gitlab.com/bark-lang/barki/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/barki/evaluator/builtins/modules"
	"gitlab.com/bark-lang/barki/object"
)

// Re-export helpers for use in this package
var (
	NULL  = helpers.NULL
	TRUE  = helpers.TRUE
	FALSE = helpers.FALSE
)

// newError is a convenience wrapper for helpers.NewError
func newError(format string, a ...interface{}) *object.Error {
	return helpers.NewError(format, a...)
}

// nativeBoolToBooleanObject is a convenience wrapper for helpers.NativeBoolToBooleanObject
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	return helpers.NativeBoolToBooleanObject(input)
}

// GetAll returns all builtin functions (global namespace + modules)
func GetAll() map[string]*object.Builtin {
	builtins := make(map[string]*object.Builtin)

	// Register global namespace builtins
	for name, fn := range InitCore() {
		builtins[name] = fn
	}
	for name, fn := range InitNumbers() {
		builtins[name] = fn
	}
	for name, fn := range InitComparison() {
		builtins[name] = fn
	}
	for name, fn := range InitArrays() {
		builtins[name] = fn
	}
	for name, fn := range InitDataStructures() {
		builtins[name] = fn
	}
	for name, fn := range InitControlFlow() {
		builtins[name] = fn
	}
	for name, fn := range InitTesting() {
		builtins[name] = fn
	}

	// Register module namespace builtins
	for name, fn := range modules.InitHTTP() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitFile() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitEnv() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitJSON() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitTime() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitBase64() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitURL() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitRegex() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitDir() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitSecurity() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitCrypto() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitMath() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitStr() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitArray() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitMap() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitIter() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitSQL() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitLazy() {
		builtins[name] = fn
	}
	for name, fn := range modules.InitCSV() {
		builtins[name] = fn
	}

	return builtins
}
