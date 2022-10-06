package BLC

// create genesis block
func (cli *CLI) createGenesisBlockchain(address string) {
	blc := CreateBlockchainWithGenesisBlock(address)
	defer blc.DB.Close()

	utxsSet := &UTXOSet{Blockchain: blc}

	utxsSet.RestUTXOSet()
}
