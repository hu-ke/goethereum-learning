package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	// "time"
	// "math/big"
	// "strings"
	// "github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/accounts/abi"
	// "github.com/ethereum/go-ethereum/accounts/abi/bind"
	// "github.com/ethereum/go-ethereum/crypto"
	// "github.com/ethereum/go-ethereum/ethclient"
)

type SolidityOutput struct {
	Contracts map[string]struct {
		ABI      interface{} `json:"abi"`
		Bytecode string      `json:"bin"`
	} `json:"contracts"`
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	file, handler, err := r.FormFile("solidity")
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

	if err != nil {
		http.Error(w, "Failed to compile Solidity file", http.StatusInternalServerError)
		return
	}

	response := Response{
		Code: 200,
		Msg:  "success",
		Data: ResponseData{
			ABI:      abi,
			Bytecode: bytecode,
		},
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
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

func askHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体
	var reqBody GptRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	url := "https://api.openai.com/v1/chat/completions"
	newReqBody := GptRequestBody2{
		Model: "gpt-3.5-turbo",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "system",
				Content: "You are a helpful assistant.",
			},
			{
				Role: "user",
				Content: "I will provide the content of a smart contract. Please analyze the contract and return the results in JSON format with the following three fields: explanation, vulnerabilities, and improvements. All fields are string type.\n" +
					"1. A brief explanation of the contract.\n" +
					"2. Any vulnerabilities or issues present in the contract. If there are no issues, please indicate that there are no recommendations.\n" +
					"3. If vulnerabilities or issues are identified, please provide suggested improvements. If there are no vulnerabilities, please indicate that there are no vulnerabilities.\n" +
					"The contract content is as follows: " + reqBody.Contract,
			},
		},
	}
	reqBodyBytes, err := json.Marshal(newReqBody)
	if err != nil {
		log.Fatalf("Error marshalling request body: %v", err)
	}
	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	apiKey := os.Getenv("API_KEY")
	req.Header.Set("Authorization", "Bearer "+apiKey) // 替换为实际的 API 密钥

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()
	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	// 解析 JSON 响应体
	var responseBody map[string]interface{}
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		log.Fatalf("Error unmarshalling response body: %v", err)
	}
	// 从响应中提取内容
	choices, ok := responseBody["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		log.Fatalf("Error: 'choices' array is missing or empty")
	}
	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		log.Fatalf("Error: The first element in 'choices' is not a valid map")
	}
	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		log.Fatalf("Error: 'message' field is missing or not a valid map")
	}
	content, ok := message["content"].(string)
	if !ok {
		log.Fatalf("Error: 'content' field is missing or not a valid map")
	}

	fmt.Println("content>>", content)

	// 将 JSON 字符串解析为结构体
	var contentJson map[string]interface{}
	err2 := json.Unmarshal([]byte(content), &contentJson)
	if err2 != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err2)
	}
	// 提取 content 对象中的字段
	explanation, _ := contentJson["explanation"].(string)
	vulnerabilities, _ := contentJson["vulnerabilities"].(string)
	improvements, _ := contentJson["improvements"].(string)
	fmt.Println("explanation", explanation)
	fmt.Println("vulnerabilities", vulnerabilities)
	// 生成响应体
	responseBody2 := GptResponseBody{
		Code: 200,
		Msg:  "success",
		Data: GptResponseContentBody{
			Explanation:     explanation,
			Vulnerabilities: vulnerabilities,
			Improvements:    improvements,
		},
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 编码响应体为 JSON
	if err := json.NewEncoder(w).Encode(responseBody2); err != nil {
		http.Error(w, "Failed to encode response body", http.StatusInternalServerError)
		return
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
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
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data ResponseData `json:"data"`
}

// 请求体的结构体
type GptRequestBody struct {
	Contract string `json:"contract"`
}
type GptResponseBody struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Data GptResponseContentBody `json:"data"`
}

// 响应体的结构体
type GptResponseContentBody struct {
	Explanation     string `json:"explanation"`
	Vulnerabilities string `json:"vulnerabilities"`
	Improvements    string `json:"improvements"`
}

// 请求体的结构体
type GptRequestBody2 struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
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
