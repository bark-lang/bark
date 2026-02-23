package token

// TokenType represents the type of a lexical token
type TokenType string

// Token represents a single lexical token
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// Token types
const (
	// Special tokens
	ILLEGAL = "ILLEGAL" // Unknown token
	EOF     = "EOF"     // End of file
	NEWLINE = "NEWLINE" // Newline (significant in bark)
	COMMENT = "COMMENT" // Comment starting with //

	// Identifiers and literals
	IDENT  = "IDENT"  // Identifier (variable names, function names)
	INT    = "INT"    // Integer literal (123, 1_000)
	FLOAT  = "FLOAT"  // Float literal (3.14, .5, 1.0e10)
	STRING = "STRING" // String literal ("hello" or `raw`)
	TRUE   = "TRUE"   // Boolean true
	FALSE  = "FALSE"  // Boolean false

	// Operators and delimiters
	DOT    = "." // .
	COMMA  = "," // ,
	COLON  = ":" // : (map key-value separator)
	GT     = ">" // > (link operator)
	MINUS  = "-" // - (only for negative number literals)
	PIPE   = "|" // | (union type separator)
	LBRACK = "[" // [
	RBRACK = "]" // ]
	LPAREN = "(" // (
	RPAREN = ")" // )
	LBRACE = "{" // {
	RBRACE = "}" // }

	// Keywords
	AS      = "AS"      // as
	ERROR   = "ERROR"   // error
	FN      = "FN"      // fn
	IMPORT  = "IMPORT"  // import
	INCLUDE = "INCLUDE" // include
	MFN     = "MFN"     // mfn (memoized function)
	MODULE  = "MODULE"  // module
	PUB     = "PUB"     // pub
	TYPE    = "TYPE"    // type
)

// keywords maps keyword strings to their token types
var keywords = map[string]TokenType{
	"as":      AS,
	"error":   ERROR,
	"fn":      FN,
	"import":  IMPORT,
	"include": INCLUDE,
	"mfn":     MFN,
	"module":  MODULE,
	"pub":     PUB,
	"type":    TYPE,
	"true":    TRUE,
	"false":   FALSE,
}

// LookupIdent checks if an identifier is a keyword
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// IsBuiltin checks if an identifier is a builtin function
// Note: Builtins are not tokenized differently, but this can be used for validation
func IsBuiltin(ident string) bool {
	builtins := map[string]bool{
		"abs": true, "absent?": true, "add": true, "append_file": true,
		"base64_decode": true, "base64_encode": true, "capture": true, "ceil": true,
		"clear": true, "contains?": true, "del": true, "delete_file": true,
		"dir_exists?": true, "div": true, "ends_with?": true, "entries": true,
		"env_all": true, "env_get": true, "env_get_or": true, "env_has?": true,
		"eprint": true, "eprint?": true, "eprintln": true, "eprintln?": true, "eq?": true, "err": true, "err_add_context": true, "err_context": true,
		"err_msg": true, "even?": true, "file_exists?": true, "file_info": true,
		"find": true, "find_all": true, "first": true, "floor": true,
		"format_iso8601": true, "format_time": true, "get": true, "get_or": true,
		"gte?": true, "gt?": true, "head": true, "http_get": true,
		"http_get_timeout": true, "http_post": true, "http_post_json": true,
		"http_request": true, "key_absent?": true, "key_present?": true,
		"index_of": true, "insert": true, "join": true, "keys": true,
		"last": true, "len": true, "list_dir": true, "lowercase": true,
		"lte?": true, "lt?": true, "match?": true, "match_indices": true,
		"max": true, "merge": true, "min": true, "mod": true,
		"mul": true, "neq?": true, "next": true, "not": true,
		"now": true, "now_ms": true, "odd?": true, "parse_iso8601": true,
		"parse_json": true, "parse_time": true, "pop": true, "present?": true,
		"prev": true, "print": true, "print?": true, "println": true, "println?": true, "push": true, "read_file": true, "regex": true,
		"remove": true, "repeat": true, "repeat?": true, "replace": true,
		"return": true, "return?": true, "round": true, "set": true,
		"shift": true, "size": true, "slice": true, "split": true,
		"starts_with?": true, "sub": true,
		"tail": true, "to_float": true, "to_int": true, "to_json": true,
		"to_json_pretty": true, "to_string": true, "trim": true, "unshift": true,
		"uppercase": true, "url_decode": true, "url_encode": true, "values": true,
		"write_file": true,
	}
	return builtins[ident]
}
