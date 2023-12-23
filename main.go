package main

import (
	"Blockchain/Lab01"
	"fmt"
	"strings"
)

func main() {
	blockchain := Lab01.InitBlockchain()
	blockchain.AddBlock("Fist block 1", "First block 2")
	blockchain.AddBlock("Second block")
	blockchain.AddBlock("Third block")

	for index, block := range blockchain.Blocks {
		fmt.Printf("The block at %d\n", index)
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		var data []string
		for _, tx := range block.Transactions {
			data = append(data, string(tx.Data))
		}
		fmt.Printf("Transactions: %s\n", strings.Join(data, ", "))
		fmt.Printf("Previous hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}
