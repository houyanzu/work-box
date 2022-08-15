package transfer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/houyanzu/work-box/lib/contract/standardcoin"
	"math/big"
	"testing"
)

func TestTransfer(t *testing.T) {
	hash, err := SingleTransfer2("0x0000000000000000000000000000000000000000", "0x927767f07bA44CDcC560DB5DE3c915159d309d1A", big.NewInt(10000000000))
	if err != nil {
		panic(err)
	}
	fmt.Println(hash)
}

func SingleTransfer2(token string, to string, amount *big.Int) (hash string, err error) {
	client, err := ethclient.Dial("https://bsc-dataseed1.defibit.io/")
	if err != nil {
		return
	}

	privateKey, err := crypto.HexToECDSA("6efef37df82d19ccf2f151ccb1a6b9db5909b19f7ac0d6072e6c0d26c75ea61c")
	if err != nil {
		return
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.NonceAt(context.Background(), fromAddress, nil)
	if err != nil {
		return
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(56))
	if err != nil {
		return
	}
	if token == "0x0000000000000000000000000000000000000000" {
		//tx := types.NewTransaction(nonce, common.HexToAddress(to), amount, 21000, gasPrice, nil)
		toAddress := common.HexToAddress(to)
		baseTx := &types.LegacyTx{
			To:       &toAddress,
			Nonce:    nonce,
			Value:    amount,
			Gas:      21000,
			GasPrice: gasPrice,
			Data:     nil,
		}
		tx := types.NewTx(baseTx)
		var signedTx *types.Transaction
		signedTx, err = types.SignTx(tx, types.NewEIP155Signer(big.NewInt(56)), privateKey)
		if err != nil {
			return
		}

		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			return
		}
		hash = signedTx.Hash().Hex()
		return
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(150000) // in units
	auth.GasPrice = gasPrice

	tokenCon := common.HexToAddress(token)
	tokenInstance, err := standardcoin.NewStandardcoin(tokenCon, client)
	if err != nil {
		return
	}
	tx, err := tokenInstance.Transfer(auth, common.HexToAddress(to), amount)
	if err != nil {
		return
	}
	hash = tx.Hash().Hex()
	return
}
