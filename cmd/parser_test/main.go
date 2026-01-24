package main

import (
	"fmt"
	"os"

	"gitlab.com/bark-lang/bark/lexer"
	"gitlab.com/bark-lang/bark/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: parser_test <file.bark>")
		os.Exit(1)
	}

	filename := os.Args[1]
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(content))
	p := parser.New(l)

	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Printf("  %s\n", msg)
		}
		os.Exit(1)
	}

	fmt.Printf("Successfully parsed %s\n", filename)
	fmt.Printf("Number of statements: %d\n\n", len(program.Statements))
	fmt.Println("AST:")
	fmt.Println(program.String())
}
