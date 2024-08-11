package main

import (
    "fmt"
    "log"
    "os"
    "github.com/ethereum/go-ethereum/accounts/keystore"
)

func main() {
    // 设置 keystore 的路径和加密密码
    ks := keystore.NewKeyStore("./keystore", keystore.StandardScryptN, keystore.StandardScryptP)
    password := "your-strong-password"

    // 生成一个新账户
    account, err := ks.NewAccount(password)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("New account address:", account.Address.Hex())

    // 输出生成的 keystore 文件路径
    fmt.Println("Keystore file path:", account.URL.Path)

    // 读取 keystore 文件内容
    fileContent, err := os.ReadFile(account.URL.Path)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Keystore file content:\n%s\n", string(fileContent))
}
