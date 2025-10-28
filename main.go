package main

import (
	"log"

	"go1f/pkg/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
