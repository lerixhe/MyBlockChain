//定义链
package main

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
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
func (bc *BlockChain) AddBlock(txs []*Transaction) {
	for _, tx := range txs {
		isOk := bc.VerifyTransaction(*tx)
		if !isOk {
			fmt.Println("invalid transaction in addBlock")
			return
		}
	}
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
	block := NewBlock(txs, lastHash)
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

//查找所有区块范围内，某个地址的所有可用UTXO所在的交易
//原理，1. 遍历所有交易中的所有inputs，由于inputs，找到对应交易，再根据对应的所有索引，找到对应所有outputs，进而认为该交易内的所有属于他outputs已经消耗。
func (bc *BlockChain) findUTXOTransactions(pubKeyHash []byte) []Transaction {
	var UTXOTransactions []Transaction
	spentUTXO := make(map[string][]int64)
	it := bc.NewBlockChainIterrator()
	for {
		//遍历区块
		block := it.Next()
		//遍历所有非coinbase交易，找出已花费的inputs,不统计这些交易了
		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, input := range tx.TXInputs {
					//input中的公钥hash与目标公钥hash相等，则认为这是转给此公钥hash的金额，且已经花过了
					//因为1个交易里的input是由output转化过来的，而这个交易已经发生了，故就认为这些inpput对应的output被花掉了
					if bytes.Equal(hash160(input.PubKey), pubKeyHash) {
						spentUTXO[string(input.TXID)] = append(spentUTXO[string(input.TXID)], input.Vout)
					}
				}
			}
		OUTPUTS:
			for currIndex, output := range tx.TXOutputs {
				if spentUTXO[string(tx.TXID)] != nil {
					indexs := spentUTXO[string(tx.TXID)]
					for _, index := range indexs {
						if int64(currIndex) == index {
							continue OUTPUTS
						}
					}
				}
				//if output.CanBeUnlockedWith(address) {
				if bytes.Equal(output.PubKeyHash, pubKeyHash) {
					UTXOTransactions = append(UTXOTransactions, *tx)
				}
			}

		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return UTXOTransactions
}

//返回指定地址的所有可用UTXO
//注意：这些UTXOs格式，只包含单纯的outputs信息，去掉了所在的交易信息
func (bc *BlockChain) FindUTXOs(pubKeyHash []byte) []TXOutput {
	var outputs []TXOutput
	//当前地址所有可用UTXO所在的交易
	txs := bc.findUTXOTransactions(pubKeyHash)
	for _, tx := range txs {
		for _, output := range tx.TXOutputs {
			//当前地址拥有的UTXO
			//if utxo.CanBeUnlockedWith(pubKeyHash) {
			if bytes.Equal(output.PubKeyHash, pubKeyHash) {
				outputs = append(outputs, output)
			}
		}
	}
	return outputs
}

//返回指定地址的满足一定余额要求的可用UTXO
//注意：这些UTXO格式要求包含所在交易ID和包含的outputs索引
func (bc *BlockChain) FindSuitableUTXO(pubKeyHash []byte, amount float64) (map[string][]int64, float64) {

	UTXOs := make(map[string][]int64)
	var total float64
	//获取某个地址的所有可用UTXO所在的交易列表
	validUTXOtxs := bc.findUTXOTransactions(pubKeyHash)
	//遍历交易
FIND:
	for _, tx := range validUTXOtxs {
		outputs := tx.TXOutputs
		//遍历outputs
		for index, output := range outputs {
			//if output.CanBeUnlockedWith(pubKeyHash) {
			if bytes.Equal(output.PubKeyHash, pubKeyHash) {
				if total < amount {
					total += output.Value
					UTXOs[string(tx.TXID)] = append(UTXOs[string(tx.TXID)], int64(index))
				} else {
					break FIND
				}
			}
		}
	}
	return UTXOs, total
}

// 对某个交易进行签名
// 输入：区块链对象，交易对象，私钥
func (bc *BlockChain) SignTransaction(tx *Transaction, priKey ecdsa.PrivateKey) {
	preTXs := make(map[string]Transaction)
	for _, input := range tx.TXInputs {
		preTX, err := bc.FindTransaction(input.TXID)
		CheckErr("fidtx err:", err)
		preTXs[string(input.TXID)] = *preTX
	}
	tx.Sign(&priKey, preTXs)
}

// 根据交易ID查找交易
func (bc *BlockChain) FindTransaction(txID []byte) (*Transaction, error) {
	it := bc.NewBlockChainIterrator()
	for {
		block := it.Next()
		for _, tx := range block.Transactions {
			if bytes.Equal(txID, tx.TXID) {
				return tx, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return nil, errors.New("Transaticon not found")
}

//区块链的校验交易方法
func (bc *BlockChain) VerifyTransaction(tx Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	preTXs := make(map[string]Transaction)
	for _, input := range tx.TXInputs {
		preTx, err := bc.FindTransaction(input.TXID)
		CheckErr("find err in verify:", err)
		preTXs[string(input.TXID)] = *preTx
	}
	return tx.Verify(preTXs)
}
