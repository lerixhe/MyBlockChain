package main

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
	//解锁脚本，指明可以使用某个output的条件
	ScriptSig string
}
type TXOutput struct {
	//支付给收款方的金额
	Value float64
	//锁定脚本，指定收款方的地址
	ScriptPubKey string
}
func 