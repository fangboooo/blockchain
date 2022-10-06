package BLC

import "fmt"

// get balance
func (cli *CLI) getBalance(address string) {
	fmt.Println("address:", address)

	blockchain := GetBlockchainObject()
	defer blockchain.DB.Close()
	utxoSet := UTXOSet{blockchain}

	amount := utxoSet.GetBalance(address)

	fmt.Printf("%s has %d token\n", address, amount)

}
