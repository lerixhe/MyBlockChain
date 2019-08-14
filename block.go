// 定义区块
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"
)

type Block struct {
	//版本
	Version int64
	//前区块的哈希
	PrevBlockHash []byte
	//当前区块的哈希，注：本成员不是区块内的信息
	Hash []byte
	//梅克尔根
	MerKelRoot []byte
	//时间戳
	TimeStamp int64
	//难度值
	Bits int64
	//随机值
	Nonce int64
	//交易信息
	Transactions []*Transaction
}

//将区块序列化为字节切片
func (block *Block) Serialize() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(block)
	CheckErr("encode err:", err)
	return buffer.Bytes()
}

//反序列化为区块
func DeSerialize(data []byte) *Block {
	var block Block
	if len(data) == 0 {
		return nil
	}
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	CheckErr("decode err", err)
	return &block
}

//创建新区块
//输入：交易，上一个区块的哈希
//输出：区块实体引用
func NewBlock(txs []*Transaction, prevBlockHash []byte) *Block {
	block := Block{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		//Hash TODO
		MerKelRoot:   []byte{},
		TimeStamp:    time.Now().Unix(),
		Bits:         targetBits,
		Nonce:        0,
		Transactions: txs}
	//block.SetHash()
	pow := NewProofOfWork(&block)
	block.Nonce, block.Hash = pow.Run()

	return &block
}

//创世块创建
func NewGenesisBlock(coinbase *Transaction) *Block {

	return NewBlock([]*Transaction{coinbase}, []byte{})

}

//区块的内所有交易的梅克尔根,这里使用拼接，简单模拟，没有使用真正的梅克尔树
func (block *Block) TransactionsHash() []byte {
	var txHashes [][]byte
	txs := block.Transactions
	for _, tx := range txs {
		txHashes = append(txHashes, tx.TXID)
	}
	data := bytes.Join(txHashes, nil)
	hash := sha256.Sum256(data)
	return hash[:]
}
