// 定义区块
package main

import (
	"bytes"
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
	Data []byte
}

// func (block *Block) SetHash() {
// 	tmp := [][]byte{
// 		IntToByte(block.Version),
// 		block.PrevBlockHash,
// 		block.MerKelRoot,
// 		IntToByte(block.TimeStamp),
// 		IntToByte(block.Bits),
// 		IntToByte(block.Nonce),
// 		block.Data}
// 	//2.参数为[][]byte类型，需要将block转成这个类型
// 	data := bytes.Join(tmp, []byte{})
// 	//1.参数为字节切片，需要构造这个切片
// 	hash := sha256.Sum256(data)
// 	block.Hash = hash[:]
// }

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
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := Block{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		//Hash TODO
		MerKelRoot: []byte{},
		TimeStamp:  time.Now().Unix(),
		Bits:       targetBits,
		Nonce:      0,
		Data:       []byte(data)}
	//block.SetHash()
	pow := NewProofOfWork(&block)
	block.Nonce, block.Hash = pow.Run()

	return &block
}

//创世块创建
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block!", []byte{})

}
