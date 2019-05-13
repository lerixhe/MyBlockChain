package main

import (
	"flag"
	"fmt"
	"os"
)

const usage = `
	createChain --address ADDRESSS "create a blockchain"
	addBlock --data DATA  "add a blockchain"
	printChain            "print all blocks"
`
const AddBlockCmdString = "addBlock"
const PrintChainCmdString = "printChain"
const CreateChainCmdString = "createChain"

type CLI struct {
	bc *BlockChain
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
	addBlockCmd := flag.NewFlagSet(AddBlockCmdString, flag.ExitOnError)
	createChainCmd := flag.NewFlagSet(CreateChainCmdString, flag.ExitOnError)
	printChainCmd := flag.NewFlagSet(PrintChainCmdString, flag.ExitOnError)
	//参数接受
	addBlockCmdPara := addBlockCmd.String("data", "", "block transaction info")
	createChainCmdPara := createChainCmd.String("address", "", "address info")
	switch os.Args[1] {
	case CreateChainCmdString:
		err := createChainCmd.Parse(os.Args[2:])
		CheckErr("parse err:", err)
		if createChainCmd.Parsed() {
			if *createChainCmdPara == "" {
				cli.PrintUsage()
			}
			cli.CreateChain(*createChainCmdPara)
		}
	case AddBlockCmdString:
		err := addBlockCmd.Parse(os.Args[2:])
		CheckErr("parse err:", err)
		if addBlockCmd.Parsed() {
			if *addBlockCmdPara == "" {
				cli.PrintUsage()
			}
			cli.AddBlock(*addBlockCmdPara)
		}
	case PrintChainCmdString:
		err := printChainCmd.Parse(os.Args[2:])
		CheckErr("printerr:", err)
		if printChainCmd.Parsed() {
			cli.PrintChain()
		}
	default:
		cli.PrintUsage()
	}
}
