package main

import (
    "context"
    "fmt"
    "log"
	"os"
	"math/big"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	infuraProjectID := os.Getenv("YOUR_INFURA_PROJECT_ID")
    ethereumAddress := os.Getenv("YOUR_ETHEREUM_ADDRESS")
    // 连接到以太坊节点
    client, err := ethclient.Dial("https://sepolia.infura.io/v3/" + infuraProjectID)
    if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	address := common.HexToAddress(ethereumAddress)
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(0),
		ToBlock:   nil, // nil for latest block
		Addresses: []common.Address{address},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatalf("Failed to retrieve logs: %v", err)
	}

	for _, vLog := range logs {
		fmt.Println("Tx Hash:", vLog.TxHash.Hex()) // Print the transaction hash
	}
}
