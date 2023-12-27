package main

import (
	"Blockchain/Lab01"
	"os"
)

func main() {
	defer os.Exit(0)
	cli := Lab01.CommandLine{}
	cli.Run()
}
