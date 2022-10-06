package BLC

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
	"math/big"
)

type Transaction struct {

	//1. transcation hash
	TxHash []byte
	//2. input
	Vins []*TXInput
	//3. output
	Vouts []*TXOutput
}

//创世区块创建时的Transaction
// 1. genesis transaction
func NewCoinbaseTransaction(address string) *Transaction {
	// consumption
	txInput := &TXInput{[]byte{}, -1, nil, []byte{}}

	txOutput := NewTXOutput(10, address)

	txCoinbase := &Transaction{
		TxHash: []byte{},
		Vins:   []*TXInput{txInput},
		Vouts:  []*TXOutput{txOutput},
	}

	// setting hash
	txCoinbase.HashTranscation()

	return txCoinbase
}

//2. transfer accounts transaction
func NewSimpleTransaction(from, to string, amount int, blockchain *BlockChain, txs []*Transaction) *Transaction {

	wallets, _ := NewWallets()
	wallet := wallets.WalletsMap[from]

	money, spendableUTXODic := blockchain.FindSpendableUTXOs(from, amount, txs)

	var txInputs []*TXInput
	var txOutputs []*TXOutput

	//1. consumption
	for txHash, indexArray := range spendableUTXODic {
		txHashBytes, _ := hex.DecodeString(txHash)
		for _, index := range indexArray {

			txInput := &TXInput{
				TxHash:    txHashBytes,
				Vout:      index,
				Signature: nil,
				PublicKey: wallet.PublicKey,
			}
			txInputs = append(txInputs, txInput)
		}
	}

	//2. transfer accounts
	txOutput := NewTXOutput(int64(amount), to)
	txOutputs = append(txOutputs, txOutput)

	//3. change note 找零
	txOutput = NewTXOutput(money-int64(amount), from)
	txOutputs = append(txOutputs, txOutput)

	tx := &Transaction{
		TxHash: []byte{},
		Vins:   txInputs,
		Vouts:  txOutputs,
	}

	// setting hash
	tx.HashTranscation()

	//进行数字签名
	blockchain.SignTransaction(tx, wallet.Privatekey, txs)

	return tx
}

// transfer hash
func (tx *Transaction) HashTranscation() {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	hash := sha256.Sum256(result.Bytes())

	tx.TxHash = hash[:]
}

// is coinbase transaction
func (tx *Transaction) IsCoinbaseTransaction() bool {
	return len(tx.Vins[0].TxHash) == 0 && tx.Vins[0].Vout == -1
}

func (tx *Transaction) Hash() []byte {
	txCopy := tx
	txCopy.TxHash = []byte{}
	hash := sha256.Sum256(txCopy.SerializeTransaction())
	return hash[:]
}

//Transaction 序列化

func (tx *Transaction) SerializeTransaction() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(tx); err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

//实现签名方法
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbaseTransaction() {
		return
	}

	for _, vin := range tx.Vins {
		if prevTXs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
			log.Panic(" ERROR:Previous transaction is not correct")
		}
	}
	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vins {
		prevTx := prevTXs[hex.EncodeToString(vin.TxHash)]
		txCopy.Vins[inID].Signature = nil
		txCopy.Vins[inID].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash
		txCopy.TxHash = txCopy.Hash()
		txCopy.Vins[inID].PublicKey = nil

		//签名代码
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.TxHash)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vins[inID].Signature = signature
	}
}

//拷贝一份新的Transaction用于数字签名
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []*TXInput
	var outputs []*TXOutput
	for _, vin := range tx.Vins {
		inputs = append(inputs, &TXInput{vin.TxHash, vin.Vout, nil, nil})
	}
	for _, vout := range tx.Vouts {
		outputs = append(outputs, &TXOutput{vout.Value, vout.Ripemd160Hash})
	}
	txCopy := Transaction{tx.TxHash, inputs, outputs}
	return txCopy

}

//验证交易输入的签名
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbaseTransaction() {
		return true
	}
	for _, vin := range tx.Vins {
		if prevTXs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
			log.Panic("ERROR:Previous transaction is not corrent")
		}
	}
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()
	for inID, vin := range tx.Vins {
		prevTx := prevTXs[hex.EncodeToString(vin.TxHash)]
		txCopy.Vins[inID].Signature = nil
		txCopy.Vins[inID].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash
		txCopy.TxHash = txCopy.Hash()
		txCopy.Vins[inID].PublicKey = nil
		//私钥ID
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PublicKey)
		x.SetBytes(vin.PublicKey[:(keyLen / 2)])
		y.SetBytes(vin.PublicKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.TxHash, &r, &s) == false {
			return false
		}
	}
	return true
}
