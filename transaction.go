package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math/big"
	"os"
	"strings"
)

const reward float64 = 12.5

type Transaction struct {
	//交易ID
	TXID []byte
	//流入记录列表
	TXInputs []TXInput
	//流出记录列表
	TXOutputs []TXOutput
}
type TXInput struct {
	//资产来源的交易ID
	TXID []byte
	//本笔交易在流出记录中的索引值
	Vout int64
	//解锁脚本，指明可以使用某个output的条件：（模拟）与此字符串相等
	//ScriptSig string
	Signature []byte //签名
	PubKey    []byte //公钥
	//转账目标是这个公钥所在的地址，你持有此公钥————钱是转给您的公钥的，而且你可以解开这个签名————你能花这笔钱
}
type TXOutput struct {
	//支付给收款方的金额
	Value float64
	//锁定脚本，指定收款方的地址
	//ScriptPubKey string
	PubKeyHash []byte //公钥哈希，目标收款地址
}

// //检查当前用户能都解开引用的utxo
// func (input *TXInput) CanUnlockUTXOWith(unlockData string) bool {
// 	return input.ScriptSig == unlockData
// }

//创建锁定脚本,锁定output的公钥hash为指定地址的公钥hash，完成锁定
//输入：钱包地址，output
func (output *TXOutput) lock(address string) {
	output.PubKeyHash = Address2hash160(address)
}

//设置交易ID，取哈希值作为ID
func (tx *Transaction) SetTXID() {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	CheckErr("encode tx err:", err)
	hash := sha256.Sum256(buffer.Bytes())
	tx.TXID = hash[:]
}

//创建Output:
//输入:转账金额，地址
//输出：output对象句柄
func NewTXOutput(value float64, address string) *TXOutput {
	output := TXOutput{Value: value}
	output.lock(address)
	return &output
}

//创建coinbase交易，为矿工奖励交易
func NewCoinBaseTrans(address, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("reward to %s %d btc", address, reward)
	}
	input := TXInput{
		TXID:      nil,
		Vout:      -1,
		Signature: nil,
		PubKey:    []byte(data)}
	output := NewTXOutput(reward, address)
	tx := Transaction{
		TXInputs:  []TXInput{input},
		TXOutputs: []TXOutput{*output}}
	tx.SetTXID()
	return &tx
}

//判断当前交易是不是coinbase
func (tx *Transaction) IsCoinbase() bool {
	if len(tx.TXInputs) == 1 {
		if len(tx.TXInputs[0].TXID) == 0 && tx.TXInputs[0].Vout == -1 {
			return true
		}
	}
	return false
}

//创建普通交易,send的辅助函数
//输入：转账地址，收款地址，金额，区块链操作句柄，输出：交易
// 1. 打开钱包容器， 根据转账人的地址找到对应的钱包
// 2. 根据钱包内的公钥哈希，取得满足条件的UTXO集合
// 3. 创建输入与输出
// 4. 将输入与输出打包成交易，签名
func NewTransaction(from, to string, amount float64, bc *BlockChain) *Transaction {
	ws := NewWallets()
	wallet := ws.Wallets[from]
	if wallet == nil {
		fmt.Printf("address ：%s not exist", from)
		return nil
	}
	pubKey := wallet.PublicKey
	privateKey := wallet.PrivateKey
	pubKeyHash := hash160(pubKey)

	vaidUTXOs, total := bc.FindSuitableUTXO(pubKeyHash, amount)
	//vaidUTXOs:所需要的可用的utxo集合.map[string][]int64.分别为交易id：该交易内属于某人的output的所有索引位置
	//total,所有可用utxo金额合计实际数量。条件：>=amount
	if total < amount {
		fmt.Println("not enough money!")
		os.Exit(1)
	}
	var inputs []TXInput
	var outputs []TXOutput
	//进行output到input的转换
	//遍历可用UTXOs得到outputs所在交易ID和outputs索引切片
	for txID, outputsIndexes := range vaidUTXOs {
		//遍历outoputs索引切片获得单个output的位置索引
		for _, outputIndex := range outputsIndexes {
			input := TXInput{TXID: []byte(txID), Vout: outputIndex, Signature: nil, PubKey: pubKey}
			inputs = append(inputs, input)
		}
	}
	//创建交易后的output
	output := *NewTXOutput(amount, to)
	outputs = append(outputs, output)
	//可能需要找零
	if total > amount {
		output := *NewTXOutput(total-amount, from)
		outputs = append(outputs, output)
	}
	tx := Transaction{
		TXID:      []byte{},
		TXInputs:  inputs,
		TXOutputs: outputs}
	tx.SetTXID()
	bc.SignTransaction(&tx, privateKey)
	return &tx
}

//签名函数：实现对某个交易内的所有inputs取得签名
//输入：交易、私钥、要签名的交易列表
func (tx *Transaction) Sign(priKey *ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	//除去coinbase交易
	if tx.IsCoinbase() {
		return
	}
	//1. 获得修改过得交易副本
	//2. 遍历交易副本中的inputs,得到上一交易的ID,
	//3. 根据交易ID和索引号，取得上一交易中的所需要的output，从而得到公钥哈希
	//4. 把公钥hash传给当前input的公钥字段，得到当前input的签私钥名
	//5. 遍历完成后，每个input都有了私钥签名数据。
	txCopy := tx.TrimmedCopy()
	for i, input := range txCopy.TXInputs {
		prevTX := prevTXs[string(input.TXID)]
		if len(prevTX.TXID) == 0 {
			fmt.Println("invalid transaction")
		}
		output := prevTX.TXOutputs[input.Vout]
		//input.PubKey = output.PubKeyHash 不能这样写，range出的变量与原变量的对应元素，地址不同
		txCopy.TXInputs[i].PubKey = output.PubKeyHash
		txCopy.SetTXID()
		txCopy.TXInputs[i].PubKey = nil
		//得到签名数据
		signData := txCopy.TXID
		r, s, err := ecdsa.Sign(rand.Reader, priKey, signData)
		CheckErr("sign err:", err)
		signature := append(r.Bytes(), s.Bytes()...)
		tx.TXInputs[i].Signature = signature
	}

}

//创建交易副本，但是所有input的签名和公钥先置空
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	for _, input := range tx.TXInputs {
		newInput := TXInput{input.TXID, input.Vout, nil, nil}
		inputs = append(inputs, newInput)
	}
	for _, output := range tx.TXOutputs {
		outputs = append(outputs, output)
	}
	return Transaction{tx.TXID, inputs, outputs}
}

// 定义校验交易过程
func (tx *Transaction) Verify(preTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	txCopy := tx.TrimmedCopy()
	for i, input := range tx.TXInputs {
		preTX := preTXs[string(input.TXID)]
		if len(preTX.TXID) == 0 {
			fmt.Println("invalid transaction in Verify")
		}
		txCopy.TXInputs[i].PubKey = preTX.TXOutputs[input.Vout].PubKeyHash
		txCopy.SetTXID()
		data := txCopy.TXID
		//开始准备校验
		signature := input.Signature
		r := big.Int{}
		s := big.Int{}
		r.SetBytes(signature[:len(signature)/2])
		s.SetBytes(signature[len(signature)/2:])
		pubKey := input.PubKey
		x := big.Int{}
		y := big.Int{}
		x.SetBytes(pubKey[:len(pubKey)/2])
		y.SetBytes(pubKey[len(pubKey)/2:])
		//构建一个原生publickey
		publicKeyOrigin := ecdsa.PublicKey{elliptic.P256(), &x, &y}
		//开始校验
		isOk := ecdsa.Verify(&publicKeyOrigin, data, &r, &s)
		if !isOk {
			return false
		}
	}
	return true
}

//打印交易
func (tx *Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.TXID))
	for i, input := range tx.TXInputs {
		lines = append(lines, fmt.Sprintf("       Input %d:", i))
		lines = append(lines, fmt.Sprintf("             TXID:%x", input.TXID))
		lines = append(lines, fmt.Sprintf("             Out:%d", input.Vout))
		lines = append(lines, fmt.Sprintf("             Signature:%x", input.Signature))
		lines = append(lines, fmt.Sprintf("             PubKey:%x", input.PubKey))
	}
	for i, output := range tx.TXOutputs {
		lines = append(lines, fmt.Sprintf("       Output %d:", i))
		lines = append(lines, fmt.Sprintf("             Value: %f:", output.Value))
		lines = append(lines, fmt.Sprintf("             Script %x:", output.PubKeyHash))
	}
	return strings.Join(lines, "\n")
}
