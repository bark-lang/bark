package main

import (
	"bufio"
	"fmt"
	"os"

	"gitlab.com/bark-lang/bark/evaluator"
	"gitlab.com/bark-lang/bark/lexer"
	"gitlab.com/bark-lang/bark/object"
	"gitlab.com/bark-lang/bark/parser"
)

const PROMPT = ">> "

func main() {
	fmt.Println("Bark REPL v0.0.1")
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

func printParserErrors(errors []string) {
	fmt.Println("Parser errors:")
	for _, msg := range errors {
		fmt.Printf("  %s\n", msg)
	}
}
