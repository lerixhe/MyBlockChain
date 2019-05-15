package main

import (
	"flag"
	"fmt"
	"os"
)

const usage = `
	createChain   --address ADDRESS      "create a blockchain"
	send          --from SOR_ADDRESS --to TAR_ADDRESS --amount AMOUNT      "send coins"
	getBalance    --address ADDRESS      "get balance of address"
	printChain    "print all blocks"
	newWallet     "create a new wallet"
`
const PrintChainCmdString = "printChain"
const CreateChainCmdString = "createChain"
const GetBalanceCmdString = "getBalance"
const SendCmdString = "send"
const NewWalletCmdString = "newWallet"

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
	//检测用户输入格式规范
	cli.parameterCheck()
	//捕获命令字符串，获取各个命令对象
	newWalletCmd := flag.NewFlagSet(NewWalletCmdString, flag.ExitOnError)
	sendCmd := flag.NewFlagSet(SendCmdString, flag.ExitOnError)
	createChainCmd := flag.NewFlagSet(CreateChainCmdString, flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet(GetBalanceCmdString, flag.ExitOnError)
	printChainCmd := flag.NewFlagSet(PrintChainCmdString, flag.ExitOnError)
	//设置命令对象的可接受参数
	fromCmdPara := sendCmd.String("from", "", "from_address info")
	toCmdPara := sendCmd.String("to", "", "target_address info")
	amountCmdPara := sendCmd.Float64("amount", 0, "amount info")
	createChainCmdPara := createChainCmd.String("address", "", "address info")
	getBalanceCmdPara := getBalanceCmd.String("address", "", "address info")
	//先switch 第一个参数
	switch os.Args[1] {
	case NewWalletCmdString:
		err := newWalletCmd.Parse(os.Args[2:])
		CheckErr("parse err:", err)
		if newWalletCmd.Parsed() {
			cli.CreateWallet()
		}
	case CreateChainCmdString:
		err := createChainCmd.Parse(os.Args[2:])
		CheckErr("parse err:", err)
		if createChainCmd.Parsed() {
			if *createChainCmdPara == "" {
				cli.PrintUsage()
			}
			cli.CreateChain(*createChainCmdPara)
		}
	case SendCmdString:
		err := sendCmd.Parse(os.Args[2:])
		CheckErr("parse err:", err)
		if sendCmd.Parsed() {
			if *fromCmdPara == "" || *toCmdPara == "" || *amountCmdPara <= 0 {
				cli.PrintUsage()
			}
			cli.Send(*fromCmdPara, *toCmdPara, *amountCmdPara)
		}
	case GetBalanceCmdString:
		err := getBalanceCmd.Parse(os.Args[2:])
		CheckErr("parse err:", err)
		if getBalanceCmd.Parsed() {
			if *getBalanceCmdPara == "" {
				cli.PrintUsage()
			}
			cli.GetBalance(*getBalanceCmdPara)
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
