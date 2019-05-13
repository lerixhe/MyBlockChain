//定义链
package main

import (
	"fmt"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockChain.db"
const blockBucket = "bucket"
const lastHashKey = "key"
const genesisInfo = "EThe Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type BlockChain struct {
	//blocks []*Block

	//数据库操作句柄
	db *bolt.DB
	//最后一个区块的哈希
	tail []byte
}

// InitBlockChain 初始化一个区块链.将原NewBLocChain()函数拆分为两个函数 初始化和获取句柄，以供其他函数使用
func InitBlockChain(address string) *BlockChain {
	if isDBFileExist() {
		fmt.Println("blockchain exist already,just use it!")
		os.Exit(1)
	}
	db, err := bolt.Open(dbFile, 0600, nil)
	CheckErr("createbucket err 0:", err)
	var lastHash []byte
	//以写的方式操作数据库
	err = db.Update(func(tx *bolt.Tx) error {

		coinbase := NewCoinBaseTrans(address, genesisInfo)
		//创建创世区块
		gblock := NewGenesisBlock(coinbase)
		//创建数据库
		bucket, err := tx.CreateBucket([]byte(blockBucket))
		CheckErr("createbucket err 1:", err)
		//写入区块信息，包括区块内容和区块的hash
		bucket.Put(gblock.Hash, gblock.Serialize())
		CheckErr("createbucket err 2:", err)
		bucket.Put([]byte(lastHashKey), gblock.Hash)
		CheckErr("createbucket err 3:", err)
		lastHash = bucket.Get([]byte(lastHashKey)) //1
		fmt.Println("create blockchain successfully")

		//既然1和2不论什么情况都要运行，即不管bucket是不是nil，都要取出最后一个区块的哈希值
		//那为何不摘出来只写一句呢？只写一句的话会bolt报错
		return nil
	})
	CheckErr("update err 1:", err)
	//返回构建成功的区块链的引用
	return &BlockChain{db, lastHash}
}

//判断数据库文件存在
func isDBFileExist() bool {
	_, err := os.Stat(dbFile)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

//获取区块链对象操作句柄
func GetBlockChainHandler() *BlockChain {
	//先判断有没有数据库管理文件
	if !isDBFileExist() {
		fmt.Println("please create blockchain first!")
		os.Exit(1)
	}
	//获取数据库文件操作句柄
	db, err := bolt.Open(dbFile, 0600, nil)
	CheckErr("dberror:", err)
	var lastHash []byte
	//以读的方式操作数据库
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		lastHash = bucket.Get([]byte(lastHashKey))
		return nil
	})
	CheckErr("update err 1:", err)
	//返回构建成功的区块链的引用
	return &BlockChain{db, lastHash}
}

//添加区块
func (bc *BlockChain) AddBlock(tss []*Transaction) {
	//读取上一区块的hash
	var lastHash []byte
	//利用区块链的数据库操作句柄，以只读方式取得上一区块的hash
	err := bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			os.Exit(1)
		}
		lastHash = bucket.Get([]byte(lastHashKey))
		return nil
	})
	CheckErr("read err 1:", err)
	//利用得到的hash生产区块
	block := NewBlock(tss, lastHash)
	//将此区块写入数据库
	err = bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			//没有这个数据库，则退出
			os.Exit(2)
		}
		bucket.Put(block.Hash, block.Serialize())
		CheckErr("putblock err 2:", err)
		bucket.Put([]byte(lastHashKey), block.Hash)
		CheckErr("putlasthashkey err 2:", err)
		return nil
	})
	CheckErr("update err 2:", err)
	//更新本地内存的区块
	bc.tail = block.Hash
}

// 迭代器：通过游标便利一个对象
type BlockChainIterator struct {
	currHash []byte
	db       *bolt.DB
}

//创建区块链迭代器
func (bc *BlockChain) NewBlockChainIterrator() *BlockChainIterator {
	return &BlockChainIterator{bc.tail, bc.db}
}

//迭代动作next

func (bci *BlockChainIterator) Next() *Block {
	var block *Block
	err := bci.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			return nil
		}
		data := bucket.Get(bci.currHash)
		block = DeSerialize(data)
		bci.currHash = block.PrevBlockHash
		return nil
	})
	CheckErr("next() err:", err)
	return block
}
