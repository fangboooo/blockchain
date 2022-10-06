package BLC

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"
)

// int64 to []byte
func Int64ToByte(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

// tansfer json to array
func JsonToArray(jsonStr string) []string {
	// json to []string
	var arr []string
	if err := json.Unmarshal([]byte(jsonStr), &arr); err != nil {
		log.Panic(err)
	}
	return arr
}

// byte array inversion
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
