package BLC

import "bytes"

type TXOutput struct {
	Value         int64
	Ripemd160Hash []byte //用户名,也就是公钥,Ripemd160Hash和input里面的publickey是一样的
}

//上锁
func (txOutput *TXOutput) Lock(address string) {

	publicKeyHash := Base58Decode([]byte(address)) //得到25字节的地址

	txOutput.Ripemd160Hash = publicKeyHash[1 : len(publicKeyHash)-4]

}

// 解锁,判断转账传的地址和TXoutput的地址是否吻合，即160hash，因为公钥进行256和160hash后，和地址反编码后的数据，是相同的
func (txOutput *TXOutput) UnLockScriptPubKeyWithAddress(address string) bool {

	publicKeyHash := Base58Decode([]byte(address)) //得到25字节的地址

	hash160 := publicKeyHash[1 : len(publicKeyHash)-4]

	return bytes.Compare(txOutput.Ripemd160Hash, hash160) == 0
}

func NewTXOutput(value int64, address string) *TXOutput {

	txOutput := &TXOutput{value, nil}

	//设置Riprmd160Hash，也就是上锁，
	txOutput.Lock(address)

	return txOutput
}
