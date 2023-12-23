package Lab01

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"time"
)

type Blockchain struct {
	Blocks []*Block
}

func InitBlockchain() *Blockchain {
	return &Blockchain{[]*Block{}}
}

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
}

func Int64ToBytes(n int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(n))
	return b
}

type Transaction struct {
	Data []byte
}

func HashTransactions(txs []*Transaction) []byte {
	var hashes [][]byte
	for _, tx := range txs {
		txHash := sha256.Sum256(tx.Data)
		hashes = append(hashes, txHash[:])
	}
	combinedHash := bytes.Join(hashes, []byte{})
	hash := sha256.Sum256(combinedHash)
	return hash[:]
}

func (b *Block) SetHash() {
	data := bytes.Join([][]byte{b.PrevBlockHash, HashTransactions(b.Transactions), Int64ToBytes(b.Timestamp)}, []byte{})
	hash := sha256.Sum256(data)
	b.Hash = hash[:]
}

func CreateBlock(data []string, prevHash []byte) *Block {
	var transactions []*Transaction
	for _, item := range data {
		transactions = append(transactions, &Transaction{[]byte(item)})
	}
	block := &Block{time.Now().UnixNano(), transactions, prevHash, []byte{}}
	block.SetHash()
	return block
}

func (blockchain *Blockchain) AddBlock(data ...string) {
	var newBlock *Block
	if len(blockchain.Blocks) == 0 {
		newBlock = CreateBlock(data, []byte{})
	} else {
		prevBlock := blockchain.Blocks[len(blockchain.Blocks)-1]
		newBlock = CreateBlock(data, prevBlock.Hash)
	}
	blockchain.Blocks = append(blockchain.Blocks, newBlock)
}
