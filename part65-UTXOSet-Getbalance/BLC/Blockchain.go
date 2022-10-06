package BLC

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

// database name
const dbName = "blockchain.db"

//table name
const blockTableName = "blocks"

type BlockChain struct {
	// the  latest block hash
	Tip []byte
	// blockchain database
	DB *bolt.DB
}

// add new block to blockchain
func (blockchain *BlockChain) AddBlockToBlockchain(txs []*Transaction) {

	// update database
	err := blockchain.DB.Update(func(tx *bolt.Tx) error {
		// 1.get table
		bucket := tx.Bucket([]byte(blockTableName))

		// 2.create new block
		if bucket != nil {
			// 3.get latest block
			blockBytes := bucket.Get(blockchain.Tip)
			// deserialize
			block := DeserializateBlock(blockBytes)

			// 4.store new block
			newBlock := NewBlock(txs, block.Height+1, block.Hash)
			err := bucket.Put(newBlock.Hash, newBlock.SerializeBlock())
			if err != nil {
				log.Panic(err)
			}

			// 4.update "l"
			err = bucket.Put([]byte("l"), newBlock.Hash)
			if err != nil {
				log.Panic(err)
			}
			// 5.update Tip
			blockchain.Tip = newBlock.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// print blockchain database
func (blockchain *BlockChain) PrintChain() {

	// create interator
	blockChainIterator := blockchain.CreateIterator()
	for {
		// get current block
		block := blockChainIterator.NextIterator()

		// print block data
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("PrevBlockHash: %x\n", block.PrevBlockHash)
		fmt.Printf("Timestamp: %s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		fmt.Println("Txs:")
		for _, tx := range block.Txs {
			fmt.Printf("tx.TxHash=%x\n", tx.TxHash)
			fmt.Printf("Vins:")
			for _, in := range tx.Vins {
				fmt.Printf("{in.TxHash:%x", in.TxHash)
				fmt.Printf(", in.Vout:%d", in.Vout)
				fmt.Printf(", in.Publickey:%s}\n", in.PublicKey)
			}
			fmt.Printf("Vouts:")
			for _, out := range tx.Vouts {
				//fmt.Printf("{out.Value:%d", out.Value)
				//fmt.Printf(", out.Ripemd160Hash:%s}\n", out.Ripemd160Hash)
				fmt.Println(out.Value)
				fmt.Println(out.Ripemd160Hash)
			}
		}
		fmt.Println()

		// doing cycle
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}

	}
}

// get balance
func (blockchain *BlockChain) GetBalance(address string) int64 {
	utxos := blockchain.UnUTXOs(address, []*Transaction{})
	var amount int64
	for _, utxo := range utxos {
		amount += utxo.OutPut.Value
	}
	return amount
}

// mine new block
func (blockchain *BlockChain) MineNewBlock(from, to, amount []string) {
	fmt.Println(from)
	fmt.Println(to)
	fmt.Println(amount)

	//1. get txs
	var txs []*Transaction
	var block *Block

	for index, address := range from {
		value, _ := strconv.Atoi(amount[index])
		tx := NewSimpleTransaction(address, to[index], value, blockchain, txs)
		txs = append(txs, tx)
		fmt.Println(tx)
	}

	//设置挖矿奖励,首先是创世区块
	tx := NewCoinbaseTransaction(from[0])
	txs = append(txs, tx)

	//1.通过算法建立Transaction数组
	blockchain.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockTableName))
		if bucket != nil {
			hash := bucket.Get([]byte("l"))

			blockBytes := bucket.Get(hash)

			block = DeserializateBlock(blockBytes)
		}
		return nil
	})

	//在建立新的区块之前对txs进行验证

	_txs := []*Transaction{}

	for _, tx := range txs {
		if blockchain.VerifyTransaction(tx, _txs) == false {
			log.Panic("签名失败......")
		}
		_txs = append(_txs, tx)
	}

	//2. add new block
	block = NewBlock(txs, block.Height+1, block.Hash)

	blockchain.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockTableName))
		if bucket != nil {
			bucket.Put(block.Hash, block.SerializeBlock())

			bucket.Put([]byte("l"), block.Hash)

			blockchain.Tip = block.Hash
		}
		return nil
	})
}

// find spend transations UTXO
func (blockchain *BlockChain) FindSpendableUTXOs(from string, amount int, txs []*Transaction) (int64, map[string][]int) {
	//1. get all UTXO
	utxos := blockchain.UnUTXOs(from, txs)
	spendableUTXO := make(map[string][]int)

	//2. traverse utxos
	var value int64

	for _, utxo := range utxos {
		value += utxo.OutPut.Value

		hash := hex.EncodeToString(utxo.TxHash)
		spendableUTXO[hash] = append(spendableUTXO[hash], utxo.Index)

		if value >= int64(amount) {
			break
		}
	}

	if value < int64(amount) {
		fmt.Printf("%s's fund is not enough\n", from)
		os.Exit(1)
	}

	return value, spendableUTXO
}

// get unspent transations
func (blockchain *BlockChain) UnUTXOs(address string, txs []*Transaction) []*UTXO {

	var unUTXOs []*UTXO
	spentTxOutputs := make(map[string][]int)

	for _, tx := range txs {
		// Vouts
	work1:
		for index, out := range tx.Vouts {
			if out.UnLockScriptPubKeyWithAddress(address) {
				//if spentTxOutputs != nil {
				fmt.Println("address:", address)
				fmt.Println("spendTXOutputs:", spentTxOutputs)

				if len(spentTxOutputs) == 0 {
					utxo := &UTXO{
						TxHash: tx.TxHash,
						Index:  index,
						OutPut: out,
					}
					unUTXOs = append(unUTXOs, utxo)
				} else {

					for hash, indexArray := range spentTxOutputs {
						txHashStr := hex.EncodeToString(tx.TxHash)
						if hash == txHashStr {

							var isSpendUTXO bool

							for _, outIndex := range indexArray {
								if index == outIndex {
									isSpendUTXO = true
									continue work1
								}

							}
							if !isSpendUTXO {
								utxo := &UTXO{
									TxHash: tx.TxHash,
									Index:  index,
									OutPut: out,
								}
								unUTXOs = append(unUTXOs, utxo)
							}

						} else {
							utxo := &UTXO{
								TxHash: tx.TxHash,
								Index:  index,
								OutPut: out,
							}
							unUTXOs = append(unUTXOs, utxo)
						}
						//}
					}
				}
			}
		}
	}

	blockIterator := blockchain.CreateIterator()

	for {
		block := blockIterator.NextIterator()
		fmt.Println(block)
		fmt.Println()

		// txHash
		for i := len(block.Txs) - 1; i >= 0; i-- {

			tx := block.Txs[i]

			// Vins
			if !tx.IsCoinbaseTransaction() {
				for _, in := range tx.Vins {
					// judge if unlock
					publicKeyHash := Base58Decode([]byte(address))           // 地址反编码后，得到hash
					ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-4] //得到公钥的160Hash

					if in.UnLockRipemd160Hash(ripemd160Hash) {
						key := hex.EncodeToString(in.TxHash)
						spentTxOutputs[key] = append(spentTxOutputs[key], in.Vout)
					}
				}
			}
			// Vouts
		work2:
			for index, out := range tx.Vouts {
				if out.UnLockScriptPubKeyWithAddress(address) {
					//if spentTxOutputs != nil {
					if len(spentTxOutputs) != 0 {

						var isSpendUTXO bool

						for txHash, indexArray := range spentTxOutputs {

							for _, i := range indexArray {
								if index == i && txHash == hex.EncodeToString(tx.TxHash) {
									isSpendUTXO = true
									continue work2
								}
							}
						}
						if !isSpendUTXO {
							utxo := &UTXO{
								TxHash: tx.TxHash,
								Index:  index,
								OutPut: out,
							}
							unUTXOs = append(unUTXOs, utxo)
						}

					} else {
						utxo := &UTXO{
							TxHash: tx.TxHash,
							Index:  index,
							OutPut: out,
						}
						unUTXOs = append(unUTXOs, utxo)
					}
					//}
				}
			}
		}

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return unUTXOs
}

// get blockchain object
func GetBlockchainObject() *BlockChain {

	// open database
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var tip []byte

	err = db.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(blockTableName))

		if bucket != nil {
			tip = bucket.Get([]byte("l"))
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return &BlockChain{
		Tip: tip,
		DB:  db,
	}
}

// 1. create genesis blockchain
func CreateBlockchainWithGenesisBlock(address string) *BlockChain {
	// is database exist
	if isDBExist() {
		fmt.Println("genesis block had exist.")
		os.Exit(1)
	}
	// create or open database
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	var genesisHash []byte
	// updata blockchain
	err = db.Update(func(tx *bolt.Tx) error {

		// 1.create table
		bucket, err := tx.CreateBucket([]byte(blockTableName))
		if err != nil {
			log.Panic(err)
		}

		if bucket != nil {
			// create a coinbase transaction
			txCoinbase := NewCoinbaseTransaction(address)

			// 2.create genesis block
			genesisBlock := CreateGenesisBlock([]*Transaction{txCoinbase})

			// 3.store genesis block to table
			err := bucket.Put(genesisBlock.Hash, genesisBlock.SerializeBlock())
			if err != nil {
				log.Panic(err)
			}

			// 4.store latest block hash
			err = bucket.Put([]byte("l"), genesisBlock.Hash)
			if err != nil {
				log.Panic(err)
			}
			genesisHash = genesisBlock.Hash
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return &BlockChain{genesisHash, db}
}

// judge isExist database
func isDBExist() bool {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}
	return true
}

//实现数字签名
func (blockchain *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey, txs []*Transaction) {

	if tx.IsCoinbaseTransaction() {
		return
	}
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vins {
		prevTX, err := blockchain.FindTransaction(vin.TxHash, txs)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.TxHash)] = prevTX
	}
	tx.Sign(privKey, prevTXs)

}

func (blockchain *BlockChain) FindTransaction(ID []byte, txs []*Transaction) (Transaction, error) {

	for _, tx := range txs {
		if bytes.Compare(tx.TxHash, ID) == 0 {
			return *tx, nil
		}
	}

	bci := blockchain.CreateIterator()
	for {
		block := bci.NextIterator()
		for _, tx := range block.Txs {
			if bytes.Compare(tx.TxHash, ID) == 0 {
				return *tx, nil
			}
		}
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp((&hashInt)) == 0 {
			break
		}
	}
	return Transaction{}, nil
}

//验证数字签名，
func (bc *BlockChain) VerifyTransaction(tx *Transaction, txs []*Transaction) bool {

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vins {
		prevTX, err := bc.FindTransaction(vin.TxHash, txs)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.TxHash)] = prevTX
	}

	return tx.Verify(prevTXs)
}

//[string]*TXOutputs
func (blc *BlockChain) FindUTXOMap() map[string]*TXOutputs {
	blcIterator := blc.CreateIterator()
	//用来存储已经消费的UTXO信息
	spentableUTXOsMap := make(map[string][]*TXInput)
	utxoMaps := make(map[string]*TXOutputs)
	for {
		block := blcIterator.NextIterator()
		for i := len(block.Txs) - 1; i >= 0; i-- {
			txOutputs := &TXOutputs{[]*UTXO{}} //传一个空的
			tx := block.Txs[i]
			//如果是coinbase交易，则跳过
			if tx.IsCoinbaseTransaction() == false {
				for _, txInput := range tx.Vins {
					txHash := hex.EncodeToString(txInput.TxHash)
					spentableUTXOsMap[txHash] = append(spentableUTXOsMap[txHash], txInput)

				}
			}
			txHash := hex.EncodeToString(tx.TxHash)

		WorkOutLoop:
			for index, out := range tx.Vouts {

				txInputs := spentableUTXOsMap[txHash]

				if len(txInputs) > 0 {
					isSpent := false
					for _, in := range txInputs {
						outPublicKey := out.Ripemd160Hash
						inPublicKey := in.PublicKey
						if bytes.Compare(outPublicKey, Ripemd160Hash(inPublicKey)) == 0 {
							if index == in.Vout {
								isSpent = true
								continue WorkOutLoop
							}
						}
					}
					if isSpent == false {
						utxo := &UTXO{tx.TxHash,index,out}
						txOutputs.UTXOs = append(txOutputs.UTXOs, utxo)
					}
				} else {
					utxo := &UTXO{tx.TxHash,index,out}
					txOutputs.UTXOs = append(txOutputs.UTXOs, utxo)
				}
			}
			//设置键值对
			utxoMaps[txHash] = txOutputs
		}
		//找到最初的创世区块后，退出
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return utxoMaps
}
