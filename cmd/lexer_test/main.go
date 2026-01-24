package main

import (
	"fmt"
	"os"
	"strings"

	"gitlab.com/bark-lang/bark/lexer"
	"gitlab.com/bark-lang/bark/token"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/lexer_test/main.go <file.bark>")
		os.Exit(1)
	}

	filename := os.Args[1]
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(content))

	fmt.Printf("Tokenizing: %s\n", filename)
	fmt.Println(strings.Repeat("=", 80))

	for {
		tok := l.NextToken()

		// Skip newlines for cleaner output
		if tok.Type == token.NEWLINE {
			continue
		}

		if tok.Type == token.EOF {
			break
		}

		// Truncate long literals
		literal := tok.Literal
		if len(literal) > 30 {
			literal = literal[:27] + "..."
		}

		fmt.Printf("Line %2d: %-10s %q\n", tok.Line, tok.Type, literal)
	}
}
