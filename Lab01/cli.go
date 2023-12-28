package Lab01

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" createblockchain -address ADDRESS creates a blockchain and sends genesis reward to address")
	fmt.Println(" getbalance -address ADDRESS - get the balance for an address")
	fmt.Println(" print - Prints the blocks in the chain")
	fmt.Println(" send - from FROM -to TO -amount AMOUNT - Send amount of coins")
	fmt.Println(" testmerkletree - run an example scenario to test merkle tree functions")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) printChain() {
	chain := ContinueBlockchain("")
	defer chain.Database.Close()
	iter := chain.Iterator()

	for {
		block := iter.Next()

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

func (cli *CommandLine) createBlockChain(address string) {
	chain := InitBlockchain(address)
	chain.Database.Close()
	fmt.Println("Finished!")
}

func (cli *CommandLine) getBalance(address string) {
	chain := ContinueBlockchain(address)
	defer chain.Database.Close()

	balance := 0
	UTXOs := chain.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int) {
	chain := ContinueBlockchain(from)
	defer chain.Database.Close()

	tx := NewTransaction(from, to, amount, chain)
	chain.AddBlock([]*Transaction{tx})
	fmt.Println("Success!")
}

func (cli *CommandLine) testMerkleTreeScenario() {
	type transaction struct {
		from  string
		to    string
		value string
	}

	var hashTx = func(t transaction) []byte {
		h := sha256.New()
		h.Write([]byte(fmt.Sprintf("%v", t)))
		return h.Sum(nil)
	}

	var printTx = func(t transaction, idx int) {
		fmt.Printf("Transaction %d : from: %s to: %s value: %s \n", idx, t.from, t.to, t.value)
	}

	trx1 := transaction{from: "mike", to: "bob", value: "100"}
	trx2 := transaction{from: "bob", to: "douglas", value: "250"}
	trx3 := transaction{from: "alice", to: "john", value: "100"}
	trx4 := transaction{from: "join", to: "mike", value: "500"}

	printTx(trx1, 1)
	printTx(trx2, 2)
	printTx(trx3, 3)
	printTx(trx4, 4)

	data := [][]byte{
		hashTx(trx1),
		hashTx(trx2),
		hashTx(trx3),
		hashTx(trx4),
	}

	// Create and verify the tree.
	merkleTree := NewMerkleTree(data, DefaultShaHasher)

	// Getting the proof of the first transaction and verify it.
	proof, idxs, err := merkleTree.GetProof(hashTx(trx1))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Verify proof of trx1: %+v \n", trx1)
	p := merkleTree.VerifyProof(hashTx(trx1), proof, idxs, DefaultShaHasher)
	fmt.Println("Proof integrity: ", p)

	// Modifying the first transaction to send money to other one.
	trx5 := transaction{from: "mike", to: "douglas", value: "10000"}
	printTx(trx5, 5)
	merkleTree.Leaves[0].Data = hashTx(trx5)
	// try verify the integrity of the tree after the modification
	fmt.Println("Tree integrity after modified: ", merkleTree.Verify())
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	testMerkleTreeCmd := flag.NewFlagSet("testmerkletree", flag.ExitOnError)

	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "testmerkletree":
		err := testMerkleTreeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if testMerkleTreeCmd.Parsed() {
		cli.testMerkleTreeScenario()
	}
}
