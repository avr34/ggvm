package main

import (
	"fmt"
	"flag"
	"log"
	"os"

	"github.com/avr34/ggvm/internal/vm"
	"github.com/avr34/ggvm/internal/vm/tokenizer"
	// "github.com/avr34/ggvm/internal/vm/core"
)

func main() {
	printtok := flag.Bool("pt", false, "print tokens")
	flag.Parse()
	
	args := flag.Args()
	if len(args) > 0 {
		file := flag.Arg(0)
		bytes, _ := os.ReadFile(file)

		source := string(bytes)

		tokenlistptr, err := tokenizer.Tokenize(source, file)
		if err != nil {
			log.Fatal(err)
		}

		state := vm.Init(*tokenlistptr)

		err = state.Run()
		if err != nil {
			log.Fatal(err)
		}
		
		if *printtok {
			fmt.Println("\nPrinting tokens")
			state.Tokens.Print()
		}
	}


	// state.Stack.Print()
}
