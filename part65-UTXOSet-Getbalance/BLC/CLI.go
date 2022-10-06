package BLC

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type CLI struct{}

// print how to use cli
func (cli *CLI) printUsage() {
	fmt.Println("\nHere is a usage...")

	fmt.Println("\taddresslist --输出所有钱包地址")
	fmt.Println("\tcreatewallet --创建钱包")
	fmt.Println("\tgetbalance -address ADDRESS")
	fmt.Println("\tcreateblockchain -address ADDRESS")
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT")
	fmt.Println("\tprintchain")
	fmt.Println("\ttest -- 测试")
}

// judge args is valid
func (cli *CLI) isValidArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// cli run
func (cli *CLI) Run() {
	testCmd := flag.NewFlagSet("test", flag.ExitOnError)
	addresslistCmd := flag.NewFlagSet("addresslist", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	sendBlockCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

	flagFrom := sendBlockCmd.String("from", "", "transfer from")
	flagTo := sendBlockCmd.String("to", "", "transfer to")
	flagAmount := sendBlockCmd.String("amount", "", "transfer amount")

	flagCreateBlockchainAddress := createBlockchainCmd.String("address", "genesis block address...", "genesis block address")

	flagGetBalanceWithAddress := getBalanceCmd.String("address", "", "get balance")

	cli.isValidArgs()

	switch os.Args[1] {
	case "test":
		err := testCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "addresslist":
		err := addresslistCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	default:
		cli.printUsage()
		os.Exit(1)
	}

	if sendBlockCmd.Parsed() {
		if *flagFrom == "" || *flagTo == "" || *flagAmount == "" {
			cli.printUsage()
			os.Exit(1)
		}

		from := JsonToArray(*flagFrom)
		to := JsonToArray(*flagTo)
		amount := JsonToArray(*flagAmount)
		for index, fromAddress := range from {
			if IsValidForAddress([]byte(fromAddress)) == false || IsValidForAddress([]byte(to[index])) == false {
				fmt.Print("地址无效......")
				cli.printUsage()
				os.Exit(1)
			}
		}
		cli.send(from, to, amount)
	}

	if createBlockchainCmd.Parsed() {
		if IsValidForAddress([]byte(*flagCreateBlockchainAddress)) == false {
			fmt.Println("地址不能为空......")
			cli.printUsage()
			os.Exit(1)
		}

		cli.createGenesisBlockchain(*flagCreateBlockchainAddress)
	}

	if getBalanceCmd.Parsed() {
		if IsValidForAddress([]byte(*flagGetBalanceWithAddress)) == false {
			fmt.Println("地址无效......")
			cli.printUsage()
			os.Exit(1)
		}
		cli.getBalance(*flagGetBalanceWithAddress)
	}

	if addresslistCmd.Parsed() {
		//输出所有区块的数据
		cli.addressList()
	}

	if printChainCmd.Parsed() {
		//输出所有区块的数据
		cli.printchain()
	}
	if testCmd.Parsed() {
		//输出所有区块的数据
		fmt.Println("测试......")
		cli.TestMethod()
	}
	if createWalletCmd.Parsed() {
		//创建钱包
		cli.createWallet()
	}
}

//18hD41rKB3L1EQ9WxEa9UeZVdfjmSPrway
