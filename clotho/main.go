package main

import (
	"fmt"
	"os"

	"github.com/julesChu12/fly/clotho/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
