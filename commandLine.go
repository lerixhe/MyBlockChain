package main

import (
	"fmt"
	"os"
)

const usage = `
	addBlock --data DATA  "add a blockchain"
	printChain            "print all blocks"
`

type CLI struct {
	bc *BlockChain
}

func (cli *CLI) AddBlock(data string) {
	cli.bc.AddBlock(data)
}
func (cli *CLI) PrintChain() {
	//打印数据
}
func (cli *CLI) PrintUsage() {
	//打印帮助
	fmt.Println(usage)
	os.Exit(1)
}
func (cli *CLI) parameterCheck() {
	if len(os.Args) < 2 {
		fmt.Println("invalid input!")
		cli.PrintUsage()
	}
}
func (cli *CLI) Run() {
	cli.parameterCheck()
}
