package main

import (
    "context"
    "crypto/ecdsa"
    "fmt"
    "log"
    "math/big"
	// "os"
    // "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    // "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
	"github.com/yourname/eth-balance-checker/dest"
)

func main() {
	infuraProjectID := os.Getenv("YOUR_INFURA_PROJECT_ID")
	youPrivateKey := os.Getenv("YOUR_PRIVATE_KEY")
    // 连接到以太坊节点
    client, err := ethclient.Dial("https://sepolia.infura.io/v3/" + infuraProjectID)
    if err != nil {
        log.Fatal(err)
    }

    // 私钥，用于部署合约的账户
    privateKey, err := crypto.HexToECDSA(youPrivateKey)
    if err != nil {
        log.Fatal(err)
    }

    // 获取公钥和账户地址
    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        log.Fatal("error casting public key to ECDSA")
    }

    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Println(fromAddress)
    // 获取nonce
    nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        log.Fatal(err)
    }

    // 设置Gas价格和Gas限制
    gasPrice, err := client.SuggestGasPrice(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    // 创建部署的交易
    auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(11155111)) // 1表示Mainnet,11155111表示sepolia
    if err != nil {
        log.Fatal(err)
    }
    auth.Nonce = big.NewInt(int64(nonce))
    auth.Value = big.NewInt(0)      // 不附带以太币
    auth.GasLimit = uint64(3000000) // 设定Gas限制
    auth.GasPrice = gasPrice

    // 部署合约
	input := "1.0"
    address, tx, _, err := store.DeployStore(auth, client, input)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("合约部署地址:", address.Hex())
    fmt.Println("交易哈希:", tx.Hash().Hex())
}
