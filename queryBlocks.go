package main

import (
    "context"
    "fmt"
    "log"
	"os"
    "github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	infuraProjectID := os.Getenv("YOUR_INFURA_PROJECT_ID")
    // ethereumAddress := os.Getenv("YOUR_ETHEREUM_ADDRESS")
    // 连接到以太坊节点
    client, err := ethclient.Dial("https://mainnet.infura.io/v3/" + infuraProjectID)
    if err != nil {
        log.Fatalf("Failed to connect to the Ethereum client: %v", err)
    }

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
	log.Fatal(err)
	}

	fmt.Println(header.Number.String()) // 5671744
}
