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
	p.SetFile(filename)

	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			fmt.Print(err.FormatError())
		}
		os.Exit(1)
	}

	fmt.Printf("Successfully parsed %s\n", filename)
	fmt.Printf("Number of statements: %d\n\n", len(program.Statements))
	fmt.Println("AST:")
	fmt.Println(program.String())
}
