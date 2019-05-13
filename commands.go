package main

import "fmt"

func (cli *CLI) AddBlock(data string) {
	//bc := GetBlockChainHandler()
	//bc.AddBlock(data)
	//cli.bc.AddBlock(data)
}
func (cli *CLI) PrintChain() {
	//打印数据
	bc := GetBlockChainHandler()
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
func (cli *CLI) CreateChain(addr string) {
	bc := InitBlockChain(addr)
	defer bc.db.Close()
}
