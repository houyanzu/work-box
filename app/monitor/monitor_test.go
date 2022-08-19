package monitor

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/houyanzu/work-box/config"
	"github.com/houyanzu/work-box/lib/mylog"
	"testing"
)

func TestName(t *testing.T) {
	logTransferSig := []byte("Transfer(address,address,uint256)")
	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)
	fmt.Println(len(logTransferSigHash.Hex()))
}

func TestListen(t *testing.T) {
	err := config.ParseConfigByFile("D:\\work\\gowork\\work-box\\config.json")
	if err != nil {
		panic(err)
	}

	err = mylog.Init("testListen.log")
	if err != nil {
		panic(err)
	}

	logss, err := Listen("0x703a195ddbc179aaf662f1954644fd7dc2f09e5d")
	if err != nil {
		panic(err)
	}

	fmt.Println("start")
	for {
		select {
		case logv := <-*logss:
			fmt.Println(logv.BlockNumber, logv.BlockHash)
			mylog.Write(logv.BlockNumber, logv.BlockHash)
		}
	}
}
