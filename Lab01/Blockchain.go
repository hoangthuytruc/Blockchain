package Lab01

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
	"time"
)

const (
	dbPath = "/tmp/blocks"
)

type Blockchain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockchainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func InitBlockchain() *Blockchain {
	var lastHash []byte
	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	Handle(err)
	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found")
			firstBlock := CreateBlock([]string{"First block"}, []byte{})
			fmt.Println("First block proved")
			err = txn.Set(firstBlock.Hash, firstBlock.Serialize())
			Handle(err)
			err = txn.Set([]byte("lh"), firstBlock.Hash)

			lastHash = firstBlock.Hash

			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			Handle(err)
			lastHash, err = item.Value()
			return err
		}
	})
	Handle(err)
	blockchain := Blockchain{lastHash, db}
	return &blockchain
}

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
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
	block := &Block{
		time.Now().UnixNano(),
		transactions,
		prevHash,
		[]byte{},
		0,
	}
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Nonce = nonce
	block.Hash = hash

	return block
}

func (blockchain *Blockchain) AddBlock(data ...string) {
	var lastHash []byte

	err := blockchain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()

		return err
	})
	Handle(err)

	newBlock := CreateBlock(data, lastHash)

	err = blockchain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		blockchain.LastHash = newBlock.Hash

		return err
	})
	Handle(err)
}

func (blockchain *Blockchain) Iterator() *BlockchainIterator {
	iter := &BlockchainIterator{blockchain.LastHash, blockchain.Database}

	return iter
}

func (iter *BlockchainIterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		encodedBlock, err := item.Value()
		block = Deserialize(encodedBlock)

		return err
	})
	Handle(err)

	iter.CurrentHash = block.PrevBlockHash

	return block
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(b)
	Handle(err)
	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	Handle(err)
	return &block
}

func Handle(err error) {
	if err != nil {
		log.Panicln(err)
	}
}
