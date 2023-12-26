package main

import (
	"Blockchain/Lab01"
	"os"
)

func main() {
	defer os.Exit(0)
	blockchain := Lab01.InitBlockchain()
	defer blockchain.Database.Close()

	cli := Lab01.CommandLine{blockchain}
	cli.Run()
}
