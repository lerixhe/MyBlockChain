//定义链
package main

import ()

type BlockChain struct {
	blocks []*Block
}

func NewBlockChain() *BlockChain {
	block := NewGenesisBlock()
	return &BlockChain{blocks: []*Block{block}}
}
func (bc *BlockChain) AddBlock(data string) {
	preblockhash := bc.blocks[len(bc.blocks)-1].Hash
	//1.需要获取上一个区块的hash
	block := NewBlock(data, preblockhash)
	//bc.blocks[len(bc.blocks)] = block
	bc.blocks = append(bc.blocks, block)
}
