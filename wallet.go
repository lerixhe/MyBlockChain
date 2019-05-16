package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"
)

const walletFileName = "wallets.dat"

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
//Bitcoin地址的生成公式：base58([prefix] + [hash160(pubkey)] + [checksum])
//过程讲解参考：https://www.jianshu.com/p/8d298e10e514   https://blog.csdn.net/jeason29/article/details/51576659
func (wallet *Wallet) GetAddress() string {
	fmt.Println("generating wallet address,please wait...")
	//取得公钥hash
	publicKeyBytes := sha256.Sum256(wallet.PublicKey)
	ripemd := ripemd160.New()
	ripemd.Write(publicKeyBytes[:])
	publicKeyHash160 := ripemd.Sum(nil)
	fmt.Printf("publicKeyHash160:%x\n", publicKeyHash160)
	//取得payload
	payload := append([]byte{0x00}, publicKeyHash160...)
	fmt.Printf("payload:         %x\n", payload)
	//取得校验序列
	temphash := sha256.Sum256(payload)
	checkSum := sha256.Sum256(temphash[:])
	fmt.Printf("checkSum:        %x\n", checkSum)
	//获得前4位校验码：
	checkCode := checkSum[:4]
	fmt.Printf("checkCode:       %x\n", checkCode)
	//获得新payload
	newpayload := append(payload, checkCode...)
	fmt.Printf("newpayload:      %x\n", newpayload)
	//获得地址
	address := base58.Encode(newpayload)
	fmt.Println("address:        ", address)
	return address
}

//定义钱包容器
type Wallets struct {
	Wallets map[string]*Wallet
}

//创建空钱包容器,先尝试从文件中读取钱包容器
func NewWallets() *Wallets {
	ws := &Wallets{make(map[string]*Wallet)}
	ws.LoadWallets()
	return ws
}

//在钱包容器内创建钱包,把钱包容器保存到本地，返回新钱包地址
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := wallet.GetAddress()
	ws.Wallets[address] = wallet
	ws.SaveWallets()
	return address
}

//保存钱包容器数据到本地
func (ws *Wallets) SaveWallets() {
	var buff bytes.Buffer
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(&ws)
	CheckErr("encode wallets err:", err)
	//写入文件
	err = ioutil.WriteFile(walletFileName, buff.Bytes(), 0644)
	CheckErr("write file err:", err)
}

// 从本地数据中加载钱包容器
func (ws *Wallets) LoadWallets() {
	//1.判断数据文件是否存在
	_, err := os.Stat(walletFileName)
	if os.IsNotExist(err) {
		return
	}
	//2.读取文件
	buf, err := ioutil.ReadFile(walletFileName)
	CheckErr("readfile err:", err)
	//3.解码
	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewBuffer(buf))
	decoder.Decode(&wallets)
	//4.传出
	//ws = &wallets
	ws.Wallets = wallets.Wallets
}

//获取钱包容器内的所有钱包地址
func (ws *Wallets) GetAllAddresses() []string {
	var addresses []string
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}
