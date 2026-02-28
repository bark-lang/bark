package main

import (
	"bufio"
	"fmt"
	"os"

	"gitlab.com/bark-lang/barki/evaluator"
	"gitlab.com/bark-lang/barki/lexer"
	"gitlab.com/bark-lang/barki/object"
	"gitlab.com/bark-lang/barki/parser"
)

const PROMPT = ">> "

func main() {
	fmt.Println("bark REPL v0.0.1")
	fmt.Println("Type 'exit' to quit")
	fmt.Println()

	env := object.NewEnvironment()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(PROMPT)
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if line == "exit" {
			break
		}

		if line == "" {
			continue
		}

		l := lexer.New(line)
		p := parser.New(l)
		p.SetFile("<repl>")

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(p.Errors())
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			fmt.Println(evaluated.Inspect())
		}
	}
}

func printParserErrors(errors []*parser.ParseError) {
	for _, err := range errors {
		fmt.Print(err.FormatError())
	}
}
