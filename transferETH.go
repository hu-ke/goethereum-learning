package main

import (
    "context"
    "fmt"
    "log"
	"os"
	"math/big"
    // "github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/core/types"
)

func main() {
	infuraProjectID := os.Getenv("YOUR_INFURA_PROJECT_ID")
    targetAddress := os.Getenv("TARGET_ETHEREUM_ADDRESS")
    youPrivateKey := os.Getenv("YOUR_PRIVATE_KEY")
    // ethereumAddress := os.Getenv("YOUR_ETHEREUM_ADDRESS")
    // 连接到以太坊节点
    client, err := ethclient.Dial("https://sepolia.infura.io/v3/" + infuraProjectID)
    // if err != nil {
    //     log.Fatalf("Failed to connect to the Ethereum client: %v", err)
    // }

	privateKey, err := crypto.HexToECDSA(youPrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(10000000000000000) // in wei (0.01 eth)
    gasLimit := uint64(21000)                // in units
    gasPrice, err := client.SuggestGasPrice(context.Background())
    if err != nil {
        log.Fatal(err)
    }
	fmt.Println(nonce, value, gasPrice, gasLimit)

	toAddress := common.HexToAddress(targetAddress)
    var data []byte
    tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

    chainID, err := client.NetworkID(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
    if err != nil {
        log.Fatal(err)
    }

    err = client.SendTransaction(context.Background(), signedTx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("tx sent: %s", signedTx.Hash().Hex())
}
