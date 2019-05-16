package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"os"
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

//检查当前用户能都解开引用的utxo
func (input *TXInput) CanUnlockUTXOWith(unlockData string) bool {
	return input.ScriptSig == unlockData
}

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
	pubKey := wallet.PublicKey
	pubKeyHash := hash160(pubKey)

	vaidUTXOs, total := bc.FindSuitableUTXO(pubKeyHash, amount)
	//vaidUTXOs:所需要的可用的utxo集合.map[string][]int64.分别为交易id：output的索引数组
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
	//创建交易后的output,由input转换而来
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

	return &tx
}
