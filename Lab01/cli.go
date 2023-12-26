package Lab01

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type CommandLine struct {
	Blockchain *Blockchain
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" add -block BLOCK_DATA - add a block to the chain")
	fmt.Println(" print - Prints the blocks in the chain")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) addBlock(data string) {
	cli.Blockchain.AddBlock(data)
	fmt.Println("Added Block!")
}

func (cli *CommandLine) printChain() {
	iter := cli.Blockchain.Iterator()
	for {
		block := iter.Next()
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		var data []string
		for _, tx := range block.Transactions {
			data = append(data, string(tx.Data))
		}
		fmt.Printf("Transactions: %s\n", strings.Join(data, ", "))
		fmt.Printf("Previous hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	blockData := addBlockCmd.String("block", "", "Block data")

	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		Handle(err)
	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		Handle(err)
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if addBlockCmd.Parsed() {
		if *blockData == "" {
			runtime.Goexit()
		}
		cli.addBlock(*blockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}
