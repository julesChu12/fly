package main

import (
	"log"
	"os"

	"github.com/julesChu12/fly/clotho/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
