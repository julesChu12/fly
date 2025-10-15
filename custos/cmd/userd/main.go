package main

import (
	"log"
	"os"

	"github.com/julesChu12/fly/custos/cmd/userd/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
