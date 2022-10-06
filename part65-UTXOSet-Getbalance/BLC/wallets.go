package BLC

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const walletFile = "wallets.dat"

//对于区块，选了一个数组将所有区块链存起来
//对于钱包，应该选字典
type Wallets struct {
	WalletsMap map[string]*Wallet
}

//创建钱包集合
func NewWallets() (*Wallets, error) {

	//判断文件是否存在,如果不存在返回err
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		//创建空钱包
		wallets := &Wallets{}
		wallets.WalletsMap = make(map[string]*Wallet)
		return wallets, err
	}
	//读数据，fileContent就是读取出来的数据
	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}
	var wallets Wallets
	gob.Register(elliptic.P256()) //声明一个wallets对象后，进行注册
	//反序列化，进行读取，将fileContent转为bytes，NewReader进行转换,bytes.NewReader(fileContent)就是最开始存储的字节数组
	decode := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decode.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	return &wallets, nil //返回*，则要带&，取地址
}

//创建一个新的钱包
func (w *Wallets) CreateNewWallet() {

	wallet := NewWallet()
	fmt.Printf("新钱包address:%s\n", wallet.GetAddress())
	w.WalletsMap[string(wallet.GetAddress())] = wallet
	//将钱包保存
	w.SaveWallets()
}

//根据地址获取钱包对象，也就是公钥和私钥
func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.WalletsMap[address]
}

////加载钱包文件
//func (ws *Wallets) LoadFromFile() error {
//	//判断文件是否存在,如果不存在返回err
//	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
//		return err
//	}
//	//读数据，fileContent就是读取出来的数据
//	fileContent, err := ioutil.ReadFile(walletFile)
//	if err != nil {
//		log.Panic(err)
//	}
//	var wallets Wallets
//	gob.Register(elliptic.P256()) //声明一个wallets对象后，进行注册
//	//反序列化，进行读取，将fileContent转为bytes，NewReader进行转换,bytes.NewReader(fileContent)就是最开始存储的字节数组
//	decode := gob.NewDecoder(bytes.NewReader(fileContent))
//	err = decode.Decode(&wallets)
//	if err != nil {
//		log.Panic(err)
//	}
//	ws.WalletsMap = wallets.WalletsMap
//	return nil
//}

//将钱包信息写入文件
func (w *Wallets) SaveWallets() {
	var content bytes.Buffer

	gob.Register(elliptic.P256()) //代表注册，目的是为了可以序列化任何类型，例如接口类型

	encode := gob.NewEncoder(&content)
	err := encode.Encode(&w)
	if err != nil {
		log.Panic(err)
	}
	//将序列化以后的数据写入到钱包文件，原来文件的数据会被覆盖
	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
