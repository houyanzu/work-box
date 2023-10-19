package app

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"log"
	"math/big"
	"testing"
)

func TestTron(t *testing.T) {
	wadRate, _ := decimal.NewFromString("998000000000000000000000000")
	wadRateB := wadRate.BigInt()
	ray := decimal.New(1, 27)
	rayB := ray.BigInt()
	wad := decimal.New(3500, 18)
	wadB := wad.BigInt()
	wadB.Mul(wadB, rayB)
	//wadRateB.Div(wadRateB, rayB)
	fmt.Println(wadB.Div(wadB, wadRateB))
}

func TestReflect(t *testing.T) {
	//logTransferSig := []byte("frob(bytes32,address,address,address,int256,int256)")
	//logTransferSigHash := crypto.Keccak256Hash(logTransferSig)
	//fmt.Println(logTransferSigHash.Hex())

	logTransferSig := []byte("slip(bytes32,address,int256)")
	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)
	fmt.Println(logTransferSigHash.Hex())

}

func TestMo(t *testing.T) {
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/d475578bcb8b4109881919997e6d660f")
	if err != nil {
		panic(err)
	}
	contractAddress := common.HexToAddress("0x35d1b3f3d7966a1dfe207aa4514c12a259a0492b")
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(17662195)),
		ToBlock:   big.NewInt(int64(17662195)),
		Addresses: []common.Address{
			contractAddress,
		},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		panic(err)
	}
	for _, log := range logs {
		fmt.Println(log.Topics[0])
	}
}

func TestKey(t *testing.T) {
	//privateKey, _ := crypto.GenerateKey()
	//privateKeyBytes := crypto.FromECDSA(privateKey)
	//privateKeyString := hexutil.Encode(privateKeyBytes)[2:]
	privateKeyString := "0000000000000000000000000000000000000000000000000000000000000100"
	privateKey, err := crypto.HexToECDSA(privateKeyString)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Println(privateKeyString)
	fmt.Println(address)
	fmt.Println(len(privateKeyString))
}
