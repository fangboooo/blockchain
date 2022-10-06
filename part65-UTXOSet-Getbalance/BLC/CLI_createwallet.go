package BLC

import (
	"fmt"
)

func (cli *CLI) createWallet() {

	wallets, _ := NewWallets() //第一次为空

	wallets.CreateNewWallet() //创建一个数据

	fmt.Println(len(wallets.WalletsMap))

}
