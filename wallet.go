package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
)

//定义钱包结构体
type Wallet struct {
	//椭圆曲线数字签名算法ECDSA
	PrivateKey ecdsa.PrivateKey
	//公钥转化为1个[]byte，方便传输
	PublicKey []byte
}

//创建钱包
func NewWallet() *Wallet {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	CheckErr("generatekey err:", err)
	publicKeyOrigin := privateKey.PublicKey
	publicKey := append(publicKeyOrigin.X.Bytes(), publicKeyOrigin.Y.Bytes()...)
	wallet := Wallet{PrivateKey: *privateKey, PublicKey: publicKey}
	return &wallet
}

//生成钱包地址：原理是使用算法将公钥转化为人类可读字符
//过程讲解参考：https://www.jianshu.com/p/8d298e10e514   https://blog.csdn.net/jeason29/article/details/51576659
func (wallet *Wallet) GetAddress() {
	//取得公钥hash
	publicKeyBytes := sha256.Sum256(wallet.PublicKey)
	ripemd := ripemd160.New()
	ripemd.Write(publicKeyBytes[:])
	publicKeyHash := ripemd.Sum(nil)
}
