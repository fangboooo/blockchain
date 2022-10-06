package BLC

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"

	"crypto/rand"
	"log"
)

const version = byte(0 * 00)
const addressChecksumLen = 4

type Wallet struct {
	//1.私钥
	Privatekey ecdsa.PrivateKey
	//2.公钥
	PublicKey []byte
}

//判断地址是否有效,也就是怎么校验
func IsValidForAddress(address []byte) bool {

	version_public_checksumBytes := Base58Decode(address) //进行一次decode解码，得到版本+公钥hash+检查码,得到25字节
	//fmt.Println(version_public_checksumBytes)
	//fmt.Printf("version_public_checksumBytes是:%x\n", version_public_checksumBytes)
	checkSumBytes := version_public_checksumBytes[len(version_public_checksumBytes)-addressChecksumLen:] //取后四位得到检查码
	//fmt.Printf("checkSumBytes是:%x\n", checkSumBytes)
	version_ripemd160 := version_public_checksumBytes[0 : len(version_public_checksumBytes)-addressChecksumLen] //取前面21个字节
	//fmt.Printf("version_ripemd160是:%x\n", version_ripemd160)
	checkBytes := CheckSum(version_ripemd160) //对前面的21个字节进行hash,可以得到检查码，如果二者相同，则地址有效
	//fmt.Printf("checkBytes是:%x\n", checkBytes)
	if bytes.Compare(checkSumBytes, checkBytes) == 0 {
		return true
	}

	return false
}

//获取钱包地址
func (w *Wallet) GetAddress() []byte {
	//1.hash160
	ripemd160Hash := Ripemd160Hash(w.PublicKey) //进行了 256以及 160 的hash,此时为20字节

	version_ripemd160Hash := append([]byte{version}, ripemd160Hash...) //添加版本号00在前面，此时为21字节

	checkSum := CheckSum(version_ripemd160Hash) //两次256hash,得到检查码，也就是前面4个字节

	bytes := append(version_ripemd160Hash, checkSum...) //在末尾添加4位检查码，得到25字节
	//数组转string
	return Base58Encode(bytes) //最后还要进行一次base58的变换,得到地址.
}

//获取检查码
func CheckSum(b []byte) []byte {
	//进行两次hash256
	firstSHA := sha256.Sum256(b)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:addressChecksumLen]
}

//获取PlicKeyHash
func Ripemd160Hash(publicKey []byte) []byte {
	//1.256
	hash256 := sha256.New()
	hash256.Write(publicKey)
	hash := hash256.Sum(nil)
	//2.160
	ripemd160 := ripemd160.New()
	ripemd160.Write(hash)
	return ripemd160.Sum(nil)
}

//创建钱包
func NewWallet() *Wallet {
	privateKey, publickey := newKeyPair()
	//fmt.Println(&privateKey)
	//fmt.Println()
	//fmt.Println(publickey)
	return &Wallet{privateKey, publickey}
}

//通过私钥产生公钥
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	//1.私钥
	curve := elliptic.P256()                              //椭圆曲线
	private, err := ecdsa.GenerateKey(curve, rand.Reader) //rand随机数,由于随机数，每次产生的私钥都是随机的
	if err != nil {
		log.Panic(err)
	}
	//2.公钥
	pubKey := append(private.PublicKey.X.Bytes(), private.Y.Bytes()...)
	return *private, pubKey
}
