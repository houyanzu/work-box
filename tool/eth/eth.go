package eth

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/houyanzu/work-box/config"
	"github.com/houyanzu/work-box/lib/contract/standardcoin"
	"github.com/houyanzu/work-box/lib/contract/unipair"
	"github.com/shopspring/decimal"
	"math/big"
	"strings"
)

func BalanceOf(token, wallet string) (balance decimal.Decimal) {
	conf := config.GetConfig()
	balance = decimal.Zero
	client, err := ethclient.Dial(conf.Eth.Host)
	if err != nil {
		return
	}

	coin, err := standardcoin.NewStandardcoin(common.HexToAddress(token), client)
	if err != nil {
		return
	}

	ba, err := coin.BalanceOf(nil, common.HexToAddress(wallet))
	if err != nil {
		return
	}
	balance = decimal.NewFromBigInt(ba, 0)
	return
}

func TokenSymbol(token string) (res string, err error) {
	conf := config.GetConfig()
	client, err := ethclient.Dial(conf.Eth.Host)
	if err != nil {
		return
	}

	coin, err := standardcoin.NewStandardcoin(common.HexToAddress(token), client)
	if err != nil {
		return
	}

	res, err = coin.Symbol(nil)
	if err != nil {
		return
	}
	return
}

func TokenName(token string) (res string, err error) {
	conf := config.GetConfig()
	client, err := ethclient.Dial(conf.Eth.Host)
	if err != nil {
		return
	}

	coin, err := standardcoin.NewStandardcoin(common.HexToAddress(token), client)
	if err != nil {
		return
	}

	res, err = coin.Name(nil)
	if err != nil {
		return
	}
	return
}

func TokenTotalSupply(token string) (res decimal.Decimal, err error) {
	conf := config.GetConfig()
	client, err := ethclient.Dial(conf.Eth.Host)
	if err != nil {
		return
	}

	coin, err := standardcoin.NewStandardcoin(common.HexToAddress(token), client)
	if err != nil {
		return
	}

	resBig, err := coin.TotalSupply(nil)
	if err != nil {
		return
	}
	res = decimal.NewFromBigInt(resBig, 0)
	return
}

func TokenDecimals(token string) (res uint8, err error) {
	conf := config.GetConfig()
	client, err := ethclient.Dial(conf.Eth.Host)
	if err != nil {
		return
	}

	coin, err := standardcoin.NewStandardcoin(common.HexToAddress(token), client)
	if err != nil {
		return
	}

	res, err = coin.Decimals(nil)
	if err != nil {
		return
	}
	return
}

func BalanceAt(wallet string) (balance decimal.Decimal, err error) {
	balance = decimal.Zero
	conf := config.GetConfig()
	client, err := ethclient.Dial(conf.Eth.Host)
	if err != nil {
		return
	}

	account := common.HexToAddress(wallet)
	ba, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return
	}

	balance = decimal.NewFromBigInt(ba, 0)
	return
}

func GetTxStatus(hash string) (status uint64, err error) {
	conf := config.GetConfig()
	client, err := ethclient.Dial(conf.Eth.Host)
	if err != nil {
		return
	}
	txHash := common.HexToHash(hash)
	tx, _, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return
	}
	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return
	}
	status = receipt.Status
	return
}

func GetUniPrice(pair, token string, amount *big.Int) (price *big.Int, err error) {
	conf := config.GetConfig()
	client, err := ethclient.Dial(conf.Eth.Host)
	if err != nil {
		return
	}

	uniPair, err := unipair.NewUnipair(common.HexToAddress(pair), client)
	if err != nil {
		return
	}

	token0, err := uniPair.Token0(nil)
	if err != nil {
		return
	}
	token1, err := uniPair.Token1(nil)
	if err != nil {
		return
	}
	token0Str := strings.ToLower(token0.Hex())
	token1Str := strings.ToLower(token1.Hex())
	tokenStr := strings.ToLower(token)

	reserves, err := uniPair.GetReserves(nil)

	if tokenStr == token0Str {
		return quote(amount, reserves.Reserve0, reserves.Reserve1), nil
	} else if tokenStr == token1Str {
		return quote(amount, reserves.Reserve1, reserves.Reserve0), nil
	}

	return nil, errors.New("wrong token")
}

func quote(amountA, reserveA, reserveB *big.Int) *big.Int {
	if amountA.Cmp(big.NewInt(0)) != 1 {
		return big.NewInt(0)
	}
	if reserveA.Cmp(big.NewInt(0)) != 1 || reserveB.Cmp(big.NewInt(0)) != 1 {
		return big.NewInt(0)
	}

	result := big.NewInt(0)
	result.Mul(amountA, reserveB).Div(result, reserveA)
	return result
}
