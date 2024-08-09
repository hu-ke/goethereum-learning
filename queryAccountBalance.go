package main

import (
    "context"
    "fmt"
    "log"
    "math/big"
	"os"
    // "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	infuraProjectID := os.Getenv("YOUR_INFURA_PROJECT_ID")
    ethereumAddress := os.Getenv("YOUR_ETHEREUM_ADDRESS")
    // 连接到以太坊节点
    client, err := ethclient.Dial("https://mainnet.infura.io/v3/" + infuraProjectID)
    if err != nil {
        log.Fatalf("Failed to connect to the Ethereum client: %v", err)
    }

    // 指定你要查询的以太坊地址
    account := common.HexToAddress(ethereumAddress)

    // 查询余额
    balance, err := client.BalanceAt(context.Background(), account, nil)
    if err != nil {
        log.Fatalf("Failed to retrieve balance: %v", err)
    }

    // 将余额从Wei转换为Ether
    etherValue := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))

    fmt.Printf("Balance of %s: %f Ether\n", account.Hex(), etherValue)
}
