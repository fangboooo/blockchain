package BLC

import (
	"bytes"
	"encoding/gob"
	"log"
)

type TXOutputs struct {
	UTXOs []*UTXO
}

// 序列化
func (txOutputs *TXOutputs) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(txOutputs); err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

//反序列化
func DeserializateTXOutputs(txOutputsBytes []byte) *TXOutputs {
	var txOutputs TXOutputs
	decoder := gob.NewDecoder(bytes.NewReader(txOutputsBytes))
	if err := decoder.Decode(&txOutputs); err != nil {
		log.Panic(err)
	}
	return &txOutputs
}
