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
	TSID []byte
	//流入记录列表
	TSInputs []TSInput
	//流出记录列表
	TSOutputs []TSOutput
}
type TSInput struct {
	//资产来源的交易ID
	TSID []byte
	//本笔交易在流出记录中的索引值
	Vout int64
	//解锁脚本，指明可以使用某个output的条件
	ScriptSig string
}
type TSOutput struct {
	//支付给收款方的金额
	Value float64
	//锁定脚本，指定收款方的地址
	ScriptPubKey string
}

//检查当前用户能都解开引用的utxo
func (input *TSInput) CanUnlockUTXOWith(unlockData string) bool {
	return input.ScriptSig == unlockData
}

//检查当前佣金光华是这个utxo的所有者
func (ouput *TSOutput) CanBeUnlockWith(unlockData string) bool {
	return ouput.ScriptPubKey == unlockData
}

//设置交易ID，取哈希值作为ID
func (ts *Transaction) SetTSID() {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(ts)
	CheckErr("encode ts err:", err)
	hash := sha256.Sum256(buffer.Bytes())
	ts.TSID = hash[:]
}

//创建coinbase交易，为矿工奖励交易
func NewCoinBaseTrans(address, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("reward to %s %d btc", address, reward)
	}
	input := TSInput{
		TSID:      []byte{},
		Vout:      -1,
		ScriptSig: data}
	output := TSOutput{
		Value:        reward,
		ScriptPubKey: address}
	ts := Transaction{
		TSInputs:  []TSInput{input},
		TSOutputs: []TSOutput{output}}
	ts.SetTSID()
	return &ts
}

//判断当前交易是不是coinbase
func (ts *Transaction) IsCoinbase() bool {
	if len(ts.TSInputs) == 1 {
		if len(ts.TSInputs[0].TSID) == 0 && ts.TSInputs[0].Vout == -1 {
			return true
		}
	}
	return false
}

//创建普通交易,send的辅助函数
func NewTransaction(from, to string, amount float64, bc *BlockChain) *Transaction {
	vaidUTXOs, total := bc.FindSuitableUTXO(from, amount)
	//vaidUTXOs:所需要的可用的utxo集合.map[string][]int64.分别为交易id：output的索引数组
	//total,所有可用utxo金额合计实际数量。>=amount
	if total < amount {
		fmt.Println("not enough money!")
		os.Exit(1)
	}
	inputs := []TSInput{}
	outputs := []TSOutput{}
	//进行output到input的转换
	//遍历可用UTXOs得到outputs所在交易ID和outputs索引切片
	for txId, outputsIndexes := range vaidUTXOs {
		//遍历outoputs索引切片获得单个output的位置索引
		for _, outputIndex := range outputsIndexes {
			input := TSInput{TSID: []byte(txId), Vout: outputIndex, ScriptSig: from}
			inputs = append(inputs, input)
		}
	}
	//创建交易后的output
	output := TSOutput{Value: amount, ScriptPubKey: to}
	outputs = append(outputs, output)
	//可能需要找零
	if total > amount {
		output := TSOutput{Value: total - amount, ScriptPubKey: to}
		outputs = append(outputs, output)
	}
	tx := Transaction{
		TSID:      []byte{},
		TSInputs:  inputs,
		TSOutputs: outputs}
	tx.SetTSID()
	return &tx
}
