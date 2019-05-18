package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"
)

func IntToByte(num int64) []byte {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.BigEndian, num)
	CheckErr("IntToByte", err)
	return buffer.Bytes()
}

func CheckErr(pos string, err error) {
	if err != nil {
		fmt.Printf("error：%s,\n at:%s", err, pos)
		os.Exit(1)
	}
}

//检查钱包地址是否符合格式要求
func CheckAddress(address string) bool {
	temp, _ := base58.Decode(address)
	checkCode := temp[len(temp)-4:]
	temphash := sha256.Sum256(temp[:len(temp)-4])
	checkSum := sha256.Sum256(temphash[:])
	//fmt.Println(checkCode, checkSum[:4])
	return bytes.Equal(checkCode, checkSum[:4])
}

//公钥生成公钥哈希：使用hash160算法
func hash160(publicKey []byte) []byte {
	publicKeyBytes := sha256.Sum256(publicKey)
	ripemd := ripemd160.New()
	ripemd.Write(publicKeyBytes[:])
	publicKeyHash160 := ripemd.Sum(nil)
	return publicKeyHash160
}

//将钱包地址转化为公钥hash
func Address2hash160(address string) []byte {
	temp, err := base58.Decode(address)
	CheckErr("decode err:", err)
	//得到地址解码数据后，拆分出公钥hash
	return temp[1 : len(temp)-4]
}
