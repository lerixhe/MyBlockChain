package main

import "fmt"

// printChain 命令
func (cli *CLI) PrintChain() {
	//打印数据
	bc := GetBlockChainHandler()
	defer bc.db.Close()
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
		fmt.Println("transactions:")
		for _, tx := range block.Transactions {
			//fmt.Printf("            %d:%x\n", i, tx.TXID)
			fmt.Println(tx.String())
		}

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
	if !CheckAddress(address) {
		fmt.Println("your address is invalid")
	}
	bc := GetBlockChainHandler()
	defer bc.db.Close()
	utxos := bc.FindUTXOs(Address2hash160(address))
	var total float64
	for _, utxo := range utxos {
		total += utxo.Value
	}
	fmt.Printf("the balance of %s is %f\n", address, total)
}

//send命令
func (cli *CLI) Send(from, to string, amount float64) {
	bc := GetBlockChainHandler()
	defer bc.db.Close()
	tx := NewTransaction(from, to, amount, bc)
	bc.AddBlock([]*Transaction{tx})
	fmt.Printf("send success,transaction ID:%x", tx.TXID)
}

//newWallet命令 创建钱包
func (cli *CLI) CreateWallet() {

	ws := NewWallets()
	address := ws.CreateWallet()
	fmt.Printf("your address:%s", address)
}

//列出钱包内所有地址
func (cli *CLI) ListAddresses() {
	ws := NewWallets()
	addresses := ws.GetAllAddresses()
	var count = 1
	for _, addr := range addresses {
		fmt.Printf("address %d: %s\n", count, addr)
		count++
	}
}

// 查找交易
func (cli *CLI) FindTX(txID []byte) {
	bc := GetBlockChainHandler()
	defer bc.db.Close()
	tx, err := bc.FindTransaction(txID)
	CheckErr("findTX err in FindTX:", err)
	_ = tx.String()
}
