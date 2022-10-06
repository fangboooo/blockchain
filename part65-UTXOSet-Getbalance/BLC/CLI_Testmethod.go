package BLC

import "fmt"

func (cli *CLI) TestMethod() {

	blockchain := GetBlockchainObject()

	defer blockchain.DB.Close()

	utxoMap := blockchain.FindUTXOMap()

	fmt.Println(utxoMap)

}
