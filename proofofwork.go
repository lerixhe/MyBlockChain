package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

const targetBits = 24

func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofOfWork{block, target}
	return pow
}

//拼装上一个区块
func (pow *ProofOfWork) PrepareData(nonce int64) []byte {
	block := pow.block
	tmp := [][]byte{
		IntToByte(block.Version),
		block.PrevBlockHash,
		block.MerKelRoot,
		IntToByte(block.TimeStamp),
		IntToByte(targetBits),
		IntToByte(nonce),
		block.Data}
	//2.参数为[][]byte类型，需要将block转成这个类型
	data := bytes.Join(tmp, []byte{})
	return data
}

//执行工作量证明，返回满足条件的随机数，和算出的哈希值
func (pow *ProofOfWork) Run() (int64, []byte) {
	var nonce int64
	var hash [32]byte
	var hashInt big.Int
	fmt.Println("Begin Mining ...")
	fmt.Printf("target hash：0000%x\n", pow.target.Bytes())
	for nonce = 1; nonce < math.MaxInt64; nonce++ {
		hash = sha256.Sum256(pow.PrepareData(nonce))
		hashInt.SetBytes(hash[:])
		//找到之后停止循环
		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("find   hash：%x\n", hash)
			fmt.Printf("find  nonce: %d\n", nonce)
			break
		}
	}
	return nonce, hash[:]

}
