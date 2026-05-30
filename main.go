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

		if *printtok {
			err = state.FirstPass()
			if err != nil{ log.Fatal(err) }
			fmt.Println("\nPrinting tokens")
			state.Tokens.Print()
		} else {
			err = state.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
	}


	// state.Stack.Print()
}
