package BLC

import (
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

//1.创建一个表来将所有的未花费输出存进来
//遍历整个数据库，读取所有未花费的UTXO，然后将所有的UTXO存储到数据库
//reset
//去遍历数据库时
//[]*TXOutputs
//
//
const utxoTableName = "utxoTableName"

type UTXOSet struct {
	Blockchain *BlockChain
}

//重置数据库表
func (utxoSet *UTXOSet) RestUTXOSet() {
	err := utxoSet.Blockchain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoTableName))
		if b != nil {
			err := tx.DeleteBucket([]byte(utxoTableName))
			if err != nil {
				log.Panic(err)
			}
		}
		b, _ = tx.CreateBucket([]byte(utxoTableName))
		if b != nil { //当删除了原先的表后，重新创建一个并遍历区块链，进行存储
			//[string]*TXOutputs
			txOutputsMap := utxoSet.Blockchain.FindUTXOMap()
			for keyHash, outs := range txOutputsMap {
				txHash, _ := hex.DecodeString(keyHash)
				b.Put(txHash, outs.Serialize())
			}
		}
		return nil
	})
	if err != nil {
		log.Panicln(err)
	}
}
func (utxoSet *UTXOSet) FindUTXOForAddress (address string) []*UTXO{
	var utxos []*UTXO
	utxoSet.Blockchain.DB.View(func(tx *bolt.Tx)error {
		// 假设表存在
		b:= tx.Bucket([]byte(utxoTableName))
		//游标
		c := b.Cursor()
		for k,v := c.First();k!= nil;k,v = c.Next(){
			fmt.Printf("key = %s,value = %s \n",k,v)//k 是hash， v是txoutput对象

			txOutputs := DeserializateTXOutputs(v)

			for _,utxo := range txOutputs.UTXOs{
				if utxo.OutPut.UnLockScriptPubKeyWithAddress(address){
					utxos = append(utxos,utxo)
				}
			}

		}
		return nil
	})
	return utxos
}

func (utxoSet *UTXOSet)GetBalance(address string) int64{
	UTXOS := utxoSet.FindUTXOForAddress(address)
	var amount int64
	for _,utxo := range UTXOS{
		amount += utxo.OutPut.Value
	}

	return amount
}