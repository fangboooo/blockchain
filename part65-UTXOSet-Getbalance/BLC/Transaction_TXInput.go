package BLC

import (
	"bytes"
)

type TXInput struct {
	//1. transaction hash
	TxHash []byte
	//2. store index in Vout
	Vout      int

	Signature []byte //数字签名

	PublicKey []byte //公钥,也就是用户名,或者，钱包里面的PublicKey,原生，没有经过160hash
}

//判断当前的消费是谁的钱
func (txInput *TXInput) UnLockRipemd160Hash(ripemd160Hash []byte) bool {

	PublicKey := Ripemd160Hash(txInput.PublicKey)

	return bytes.Compare(PublicKey, ripemd160Hash) == 0
}

////判断当前输入是否和某个输出吻合
//func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
//	//判断输出里面的公钥是否和lockingHash对应
//
//	lockingHash := HashPubKey(in.PubKey)
//
//	return bytes.Compare(lockingHash, pubKeyHash) == 0
//
//}
