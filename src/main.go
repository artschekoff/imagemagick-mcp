package main

import (
	"log"
	"os"
)

// version is injected at build time via -ldflags "-X main.version=<tag>"
var version = "dev"

func main() {
	if err := serve(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
