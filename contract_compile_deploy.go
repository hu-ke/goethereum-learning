package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"encoding/json"
	"time"
	"math/big"
	"strings"
	"github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type SolidityOutput struct {
	Contracts map[string]struct {
		ABI      interface{} `json:"abi"`
		Bytecode string      `json:"bin"`
	} `json:"contracts"`
}

// extractBetween extracts a substring between two delimiters.
func extractBetween(value, a, b string) string {
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	posLast := strings.Index(value[posFirstAdjusted:], b)
	if posLast == -1 {
		return ""
	}
	posLastAdjusted := posFirstAdjusted + posLast
	return strings.TrimSpace(value[posFirstAdjusted:posLastAdjusted])
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	file, handler, err := r.FormFile("solidity")
	fmt.Println(file)
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	// fmt.Printf("File Size: %+v\n", handler.Size)
	// fmt.Printf("MIME Header: %+v\n", handler.Header)

	filePath := "./contracts/" + handler.Filename
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	// 将上传的文件内容复制到本地文件中
	if _, err := io.Copy(out, file); err != nil {
		fmt.Println("Error saving file")
		fmt.Println(err)
		return
	}

	// Compile the Solidity file
	bytecode, abi, err := compileSolidity(handler.Filename)
	// fmt.Println(bytecode, abi, err)
	if err != nil {
		http.Error(w, "Failed to compile Solidity file", http.StatusInternalServerError)
		return
	}

	response := Response{
		Code: 200,
		Msg: "success",
		Data: ResponseData {
			ABI: abi,
			Bytecode: bytecode,
		},
	}
	jsonResponse, err := json.Marshal(response)
	if err!= nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}

	// 设置内容类型为 JSON 并返回响应
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(jsonResponse)
	// Deploy the contract
	// address, deployTime, err := deployContract(bytecode, abi)
	// if err != nil {
	// 	http.Error(w, "Failed to deploy contract", http.StatusInternalServerError)
	// 	return
	// }

	// // // Return deployment info
	// fmt.Fprintf(w, "Deployed contract at address: %s\nDeployment time: %s\n", address.Hex(), deployTime)
}

func compileSolidity(filename string) (string, string, error) {
	filePath := "./contracts/" + filename
	// 构建solc命令
	cmd := exec.Command("solc",
		"--include-path", "node_modules/",
		"--base-path", ".",
		filePath,
		"--combined-json", "abi,bin")
	// Run the command and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to execute solc: %v, output: %s", err, output)
	}

	// Parse the JSON output
	var solOutput SolidityOutput
	if err := json.Unmarshal(output, &solOutput); err != nil {
		return "", "", fmt.Errorf("failed to parse solc output: %v", err)
	}

	// Access the ABI and Bytecode
	fmt.Println(solOutput.Contracts)

	for _, data := range solOutput.Contracts {
		abi, err := json.Marshal(data.ABI)
		if err != nil {
			return "", "", fmt.Errorf("failed to marshal ABI: %v", err)
		}
		return data.Bytecode, string(abi), nil
	}
	return "", "", nil
}

func deployContract(bytecode string, abi string) (common.Address, string, error) {
	client, err := ethclient.Dial("https://sepolia.infura.io/v3/" + infuraProjectID)
	if err != nil {
		return common.Address{}, "", err
	}

	privateKey, err := crypto.HexToECDSA(os.Getenv("YOUR_PRIVATE_KEY"))

	if err != nil {
		return common.Address{}, "", err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(11155111)) // Chain ID 1 for mainnet
	if err != nil {
		return common.Address{}, "", err
	}

	var StoreMetaData = &bind.MetaData{
		ABI: abi,
		Bin: bytecode,
	}
	parsed, err := StoreMetaData.GetAbi()

	address, tx, _, err := bind.DeployContract(auth, *parsed, common.FromHex(bytecode), client)
	fmt.Println("address", address)
	fmt.Println(tx)
	if err != nil {
		return common.Address{}, "", err
	}

	deployTime := time.Now().Format(time.RFC3339)
	return address, deployTime, nil
}

func main() {
	http.HandleFunc("/api/ask-gpt", askHandler)
	http.HandleFunc("/api/upload", uploadHandler)
	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", corsMiddleware(http.DefaultServeMux))
}

// 定义数据结构
type ResponseData struct {
    ABI      string `json:"abi"`
    Bytecode string `json:"bytecode"`
}

type Response struct {
    Code int         `json:"code"`
    Msg  string      `json:"msg"`
    Data ResponseData `json:"data"`
}

// CORS Middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}