package BLC

import "fmt"

//输出所有的钱包地址
func (cli *CLI) addressList() {

	fmt.Println("打印所有的钱包地址")

	wallets, _ := NewWallets()

	for address, _ := range wallets.WalletsMap {
		fmt.Println(address)
	}

}
