package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"sepolia-block/counter"
)

func blockTest() {
	// 连接测试节点main
	client, err := ethclient.Dial("https://1rpc.io/sepolia")
	if err != nil {
		log.Fatalf("连接失败： %v", err)
	}
	defer client.Close() // 关闭

	// 指定区块号
	blockNumber := big.NewInt(1898989)

	// 区块信息main
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		log.Fatalf("区块获取失败：%v", err)
	}

	// 输出区块信息
	fmt.Printf("区块号: %d\n", block.NumberU64())
	fmt.Printf("区块哈希: %s\n", block.Hash().Hex())
	fmt.Printf("时间戳: %d\n", block.Time())
	fmt.Printf("交易数量: %d\n", len(block.Transactions()))

	privateKeyHex := os.Getenv("private_key")
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	// 查询 nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	// 设置转账参数
	toAddress := common.HexToAddress("0xEfDA589312a37aB1b0cac1f11d5b96117D31bCF9")
	value := big.NewInt(1e14) // 0.0001 ETH (1e18 = 1 ETH)
	gasLimit := uint64(21000)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 构造交易
	tx := types.NewTransaction(
		nonce,
		toAddress,
		value,
		gasLimit,
		gasPrice,
		nil,
	) //

	// 获取链 ID（Sepolia = 11155111）
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 签名交易
	signedTx, err := types.SignTx(
		tx,
		types.NewEIP155Signer(chainID),
		privateKey,
	)
	if err != nil {
		log.Fatal(err)
	}

	// 发送交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	// 输出交易哈希
	fmt.Printf("交易已发送 🎉\nTx Hash: %s\n", signedTx.Hash().Hex())

}

func main() {
	// forge build 某合约, 生成json
	// jq '.abi' out/Counter.sol/Counter.json > Counter.abi
	// jq -r '.bytecode.object' out/Counter.sol/Counter.json > Counter.bin
	// abigen \ --abi build/Counter.abi \ --bin build/Counter.bin \ --pkg counter \ --out counter.go

	client, err := ethclient.Dial("https://1rpc.io/sepolia")
	if err != nil {
		log.Fatalf("连接失败： %v", err)
	}
	defer client.Close() // 关闭

	privateKeyHex := os.Getenv("private_key")
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(
		privateKey,
		big.NewInt(11155111), // Sepolia chainID
	)
	if err != nil {
		log.Fatal(err)
	}

	// 合约地址已部署
	contractAddress := common.HexToAddress("0xe09d7Ce1107Dc37C9c20d8019DD1786Ca82F6640")
	c, err := counter.NewCounter(contractAddress, client)
	if err != nil {
		log.Fatal(err)
	}

	// 调用 inc() 修改状态
	tx, err := c.Inc(auth)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Increment transaction sent:", tx.Hash().Hex())

	// 调用自动生成 get() 读取当前计数
	num, err := c.X(&bind.CallOpts{
		Pending: true,
		From:    auth.From,
		Context: context.Background(),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Current counter value:", num)
}
