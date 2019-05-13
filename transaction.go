package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
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

func (ts *Transaction) SetTXID() {
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
	tx := Transaction{
		TSInputs:  []TSInput{input},
		TSOutputs: []TSOutput{output}}
	tx.SetTXID()
	return &tx
}

//创建普通交易
