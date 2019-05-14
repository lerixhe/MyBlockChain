package main

import "fmt"

// addblock 命令
func (cli *CLI) AddBlock(data string) {
	//bc := GetBlockChainHandler()
	//bc.AddBlock(data)
	//cli.bc.AddBlock(data)
}

// printChain 命令
func (cli *CLI) PrintChain() {
	//打印数据
	bc := GetBlockChainHandler()
	bc.db.Close()
	bci := bc.NewBlockChainIterrator()
	for {
		block := bci.Next()
		fmt.Println("———————————————————————————————————")
		fmt.Printf("Version:%d\n", block.Version)
		fmt.Printf("PrevBlockHash:%x\n", block.PrevBlockHash)
		fmt.Printf("Hash:%x\n", block.Hash)
		fmt.Printf("TimeStamp:%d\n", block.TimeStamp)
		fmt.Printf("Bits:%d\n", block.Bits)
		fmt.Printf("Nonce:%d\n", block.Nonce)
		//fmt.Printf("Data:%s\n", block.Data)
		fmt.Printf("isValid:%t\n", NewProofOfWork(block).Isvalid(block.Nonce))
		fmt.Println("___________________________________")
		fmt.Println("               ||                  ")
		if len(block.PrevBlockHash) == 0 {
			fmt.Println("print over!")
			break
		}
	}
}

// createChain命令
func (cli *CLI) CreateChain(addr string) {
	bc := InitBlockChain(addr)
	defer bc.db.Close()
}

// getbalance 命令
func (cli *CLI) GetBalance(address string) {
	bc := GetBlockChainHandler()
	defer bc.db.Close()
	utxos := bc.FindUTXO(address)
	var total float64 = 0.0
	for _, utxo := range utxos {
		total += utxo.Value
	}
	fmt.Printf("the balance od %s is %f\n", address, total)
}
