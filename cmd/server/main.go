package main

import (
	"fmt"
	"os"

	"github.com/Stolkerve/tiny-kafka/internal"
)

func main() {
	server := internal.NewServer("localhost:5678")
	err := server.ListenConections()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(-1)
	}

}
