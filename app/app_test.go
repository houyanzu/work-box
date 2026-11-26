package app

import (
	"github.com/houyanzu/work-box/tool/eth"
	"math/big"
	"testing"
)

func Test(t *testing.T) {
	account, err := eth.GetClientAndAuthStruct(1, "私钥", gasLimit, big.NewInt(0))
	if err != nil {
		return
	}
	hash, err := account.MultiMint("代币合约地址", tos, amounts)
	if err != nil {
		return
	}
}
