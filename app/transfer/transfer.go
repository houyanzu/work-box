package transfer

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	tronClient "github.com/fbsobreira/gotron-sdk/pkg/client"
	common2 "github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/houyanzu/work-box/database/models/chains"
	"github.com/houyanzu/work-box/database/models/keys"
	"github.com/houyanzu/work-box/database/models/locktransferdetails"
	"github.com/houyanzu/work-box/database/models/pwdwt"
	"github.com/houyanzu/work-box/database/models/transferdetails"
	"github.com/houyanzu/work-box/database/models/transferrecords"
	"github.com/houyanzu/work-box/lib/contract/locktransfer"
	"github.com/houyanzu/work-box/lib/contract/multitransfer"
	"github.com/houyanzu/work-box/lib/contract/standardcoin"
	crypto2 "github.com/houyanzu/work-box/lib/crypto"
	"github.com/houyanzu/work-box/lib/crypto/aes"
	"github.com/houyanzu/work-box/lib/mytime"
	"github.com/houyanzu/work-box/lib/tron"
	"github.com/houyanzu/work-box/tool/eth"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"math/big"
	"strings"
	"time"
)

var privateKeyStr string
var FromAddress string
var TronFromAddress string
var GasLimit = uint64(100000)

var CheckBalance = false

func InitTrans(priKeyCt aes.Decoder, password []byte) (e error) {
	defer func() {
		err := recover()
		if err != nil {
			pwdwt.New(nil).Wrong()
			e = errors.New("wrong password")
			return
		}
	}()
	times := pwdwt.New(nil).GetTimes()
	if times >= 5 {
		e = errors.New("locked")
		return
	}
	privateKeyByte := priKeyCt.Decode(password)
	privateKeyStr = privateKeyByte.ToString()
	pwdwt.New(nil).ResetTimes()

	privateKey, e := crypto.HexToECDSA(privateKeyStr)
	if e != nil {
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		e = errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return
	}
	FromAddress = crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	TronFromAddress, _ = tron.HexToTronAddress(FromAddress)
	return
}

func InitDBTrans(priKeyID uint, password []byte, de crypto2.Decoder) (e error) {
	defer func() {
		err := recover()
		if err != nil {
			pwdwt.New(nil).Wrong()
			e = errors.New("wrong password")
			return
		}
	}()
	times := pwdwt.New(nil).GetTimes()
	if times >= 5 {
		e = errors.New("locked")
		return
	}
	priKeyModel := keys.New(nil).InitByID(priKeyID)
	if priKeyModel.Data.ID == 0 {
		e = errors.New("priKey not exists")
		return
	}
	privateKeyStr = priKeyModel.GetPriKey(password, de)
	pwdwt.New(nil).ResetTimes()

	privateKey, e := crypto.HexToECDSA(privateKeyStr)
	if e != nil {
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		e = errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return
	}
	FromAddress = crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	TronFromAddress, _ = tron.HexToTronAddress(FromAddress)
	return
}

func Transfer(chainDBID uint, limit int, module string) (err error) {
	chain := chains.New(nil).InitByID(chainDBID)
	if !chain.Exists() {
		err = errors.New("chain not found")
		return
	}
	fromAddress := FromAddress
	if chain.Data.Name == "Tron" {
		limit = 1
		fromAddress, _ = tron.HexToTronAddress(fromAddress)
	}
	pending := transferrecords.New(nil).InitPending(chainDBID, fromAddress, module)
	if pending.Exists() {
		var status uint64
		status, err = eth.GetTxStatus(chainDBID, pending.Data.Hash)
		if err != nil {
			exTime := pending.Data.CreateTime.Add(30 * time.Minute)
			now := mytime.NewFromNow()
			if exTime.Before(now) {
				pending.SetFail()
				if pending.Data.Type == 1 {
					transferdetails.New(nil).Reset(pending.Data.ID)
				} else if pending.Data.Type == 2 {
					locktransferdetails.New(nil).Reset(pending.Data.ID)
				}

			}
			return
		}
		fmt.Println(status, "-------------")
		if status == 1 {
			pending.SetSuccess()
			if pending.Data.Type == 1 {
				transferdetails.New(nil).SetSuccess(pending.Data.ID)
			} else if pending.Data.Type == 2 {
				locktransferdetails.New(nil).SetSuccess(pending.Data.ID)
			}
		} else {
			pending.SetFail()
			if pending.Data.Type == 1 {
				transferdetails.New(nil).SetFail(pending.Data.ID)
			} else if pending.Data.Type == 2 {
				locktransferdetails.New(nil).SetFail(pending.Data.ID)
			}
		}
	}

	waitingList := transferdetails.New(nil).InitWaitingList(chainDBID, limit, module)
	length := len(waitingList.List)
	if length == 0 {
		return
	}
	if chain.Data.Name == "Tron" {
		fmt.Printf("%+v\n", waitingList.List[0])
		if waitingList.List[0].Token[:2] == "0x" && waitingList.List[0].Token != eth.EthAddress {
			waitingList.List[0].Token, _ = tron.HexToTronAddress(waitingList.List[0].Token)
		}
		opts := make([]grpc.DialOption, 0)
		opts = append(opts, grpc.WithInsecure())

		conn := tronClient.NewGrpcClient(chain.Data.Rpc)
		conn.SetTimeout(5 * time.Second)

		if err = conn.Start(opts...); err != nil {
			return
		}

		err = conn.SetAPIKey(chain.Data.ApiKey)
		if err != nil {
			return
		}

		if CheckBalance {
			balance, errr := eth.BalanceOf(chainDBID, waitingList.List[0].Token, fromAddress)
			if errr != nil {
				err = errr
				return
			}
			if balance.LessThan(waitingList.List[0].Amount) {
				err = errors.New("insufficient balance: " + waitingList.List[0].Token)
				return
			}
		}

		tronTo := waitingList.List[0].To
		if tronTo[:2] == "0x" {
			tronTo, _ = tron.HexToTronAddress(tronTo)
		}
		var tx *api.TransactionExtention
		if waitingList.List[0].Token != eth.EthAddress {
			tx, err = conn.TRC20Send(fromAddress, tronTo, waitingList.List[0].Token, waitingList.List[0].Amount.BigInt(), 15000000)
			if err != nil {
				return
			}
		} else {
			tx, err = conn.Transfer(fromAddress, tronTo, waitingList.List[0].Amount.IntPart())
			if err != nil {
				return
			}
		}
		signedTx, errr := tron.SignTx(privateKeyStr, tx.Transaction)
		if errr != nil {
			err = errr
			return
		}
		_, err = conn.Broadcast(signedTx)
		if err != nil {
			return
		}

		tr := transferrecords.New(nil)
		tr.Data.Type = 1
		tr.Data.ChainDbId = chainDBID
		tr.Data.From = fromAddress
		tr.Data.Hash = common2.BytesToHexString(tx.GetTxid())
		tr.Data.Nonce = 0
		tr.Data.Module = module
		tr.Add()

		transferdetails.New(nil).SetExec([]uint{waitingList.List[0].ID}, tr.Data.ID)
		return
	}

	tokens, tos := make([]common.Address, length), make([]common.Address, length)
	ids := make([]uint, length)
	amounts := make([]*big.Int, length)
	totalValue := big.NewInt(0)

	transferTokens := make(map[string]decimal.Decimal)
	waitingList.Foreach(func(index int, m *transferdetails.Model) {
		tokens[index] = common.HexToAddress(m.Data.Token)
		tos[index] = common.HexToAddress(m.Data.To)
		if _, ok := transferTokens[strings.ToLower(m.Data.Token)]; ok {
			transferTokens[strings.ToLower(m.Data.Token)] = transferTokens[strings.ToLower(m.Data.Token)].Add(m.Data.Amount)
		} else {
			transferTokens[strings.ToLower(m.Data.Token)] = m.Data.Amount
		}
		amount := m.Data.Amount.BigInt()
		amounts[index] = amount
		ids[index] = m.Data.ID
		if m.Data.Token == eth.EthAddress {
			totalValue.Add(totalValue, amount)
		}
	})
	if CheckBalance {
		for token, amount := range transferTokens {
			if token != eth.EthAddress {
				ba, _ := eth.BalanceOf(chainDBID, token, fromAddress)
				if ba.LessThan(amount) {
					err = errors.New("insufficient balance: " + token)
					return
				}
			}
		}
	}
	client, err := ethclient.Dial(chain.Data.Rpc)
	if err != nil {
		return
	}

	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return
	}
	nonce, err := client.NonceAt(context.Background(), common.HexToAddress(fromAddress), nil)
	if err != nil {
		return
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chain.Data.ChainID))
	if err != nil {
		return
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = totalValue                                  // in wei
	auth.GasLimit = GasLimit * uint64(len(waitingList.List)) // in units
	auth.GasPrice = gasPrice

	multiCon := common.HexToAddress(chain.Data.MultiTransferContract)
	multiTransferInstance, err := multitransfer.NewMultitransfer(multiCon, client)
	if err != nil {
		return
	}
	tx, err := multiTransferInstance.MultiTransferToken(auth, tokens, tos, amounts)
	if err != nil {
		return
	}
	hash := tx.Hash().Hex()

	tr := transferrecords.New(nil)
	tr.Data.Type = 1
	tr.Data.ChainDbId = chainDBID
	tr.Data.From = fromAddress
	tr.Data.Hash = hash
	tr.Data.Nonce = nonce
	tr.Data.Module = module
	tr.Add()

	transferdetails.New(nil).SetExec(ids, tr.Data.ID)
	return
}

func LockTransfer(chainDBID uint, module string) (err error) {
	limit := 1
	chain := chains.New(nil).InitByID(chainDBID)
	if !chain.Exists() {
		err = errors.New("chain not found")
		return
	}
	pending := transferrecords.New(nil).InitPending(chainDBID, FromAddress, module)
	if pending.Data.ID > 0 {
		var status uint64
		status, err = eth.GetTxStatus(chainDBID, pending.Data.Hash)
		if err != nil {
			//TODO:覆盖操作
			return
		}
		if status == 1 {
			pending.SetSuccess()
			if pending.Data.Type == 1 {
				transferdetails.New(nil).SetSuccess(pending.Data.ID)
			} else if pending.Data.Type == 2 {
				locktransferdetails.New(nil).SetSuccess(pending.Data.ID)
			}
		} else if status == 0 {
			pending.SetFail()
			if pending.Data.Type == 1 {
				transferdetails.New(nil).SetFail(pending.Data.ID)
			} else if pending.Data.Type == 2 {
				locktransferdetails.New(nil).SetFail(pending.Data.ID)
			}
		}
	}

	waitingList := locktransferdetails.New(nil).InitWaitingList(chainDBID, limit, module)
	length := len(waitingList.List)
	if length == 0 {
		return
	}

	waiting := locktransferdetails.New(nil).InitByData(waitingList.List[0])

	client, err := ethclient.Dial(chain.Data.Rpc)
	if err != nil {
		return
	}

	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return
	}
	nonce, err := client.NonceAt(context.Background(), common.HexToAddress(FromAddress), nil)
	if err != nil {
		return
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chain.Data.ChainID))
	if err != nil {
		return
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = GasLimit   // in units
	auth.GasPrice = gasPrice

	ltCon := common.HexToAddress(chain.Data.LockTransferContract)
	ltInstance, err := locktransfer.NewLocktransfer(ltCon, client)
	if err != nil {
		return
	}
	tx, err := ltInstance.LockTransfer(
		auth,
		common.HexToAddress(waiting.Data.Token),
		common.HexToAddress(waiting.Data.To),
		waiting.Data.Amount.BigInt(),
		big.NewInt(int64(waiting.Data.ReleaseStartTime)),
		big.NewInt(int64(waiting.Data.ReleaseCycle)),
		big.NewInt(int64(waiting.Data.ReleaseTimes)),
	)
	if err != nil {
		return
	}
	hash := tx.Hash().Hex()

	tr := transferrecords.New(nil)
	tr.Data.ChainDbId = chainDBID
	tr.Data.Type = 2
	tr.Data.From = FromAddress
	tr.Data.Hash = hash
	tr.Data.Nonce = nonce
	tr.Add()

	locktransferdetails.New(nil).SetExec([]uint{waiting.Data.ID}, tr.Data.ID)
	return
}

func SingleTransfer(chainDBID uint, token string, to string, amount *big.Int, priKey string) (hash string, nonce uint64, err error) {
	chain := chains.New(nil).InitByID(chainDBID)
	if !chain.Exists() {
		err = errors.New("chain not found")
		return
	}

	if priKey == "" {
		priKey = privateKeyStr
	} else {
		privateKey, e := crypto.HexToECDSA(priKey)
		if e != nil {
			return
		}

		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			e = errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
			return
		}
		FromAddress = crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
		TronFromAddress, _ = tron.HexToTronAddress(FromAddress)
	}
	if chain.Data.Name == "Tron" {
		opts := make([]grpc.DialOption, 0)
		opts = append(opts, grpc.WithInsecure())

		conn := tronClient.NewGrpcClient(chain.Data.Rpc)

		if err = conn.Start(opts...); err != nil {
			return
		}

		err = conn.SetAPIKey(chain.Data.ApiKey)
		if err != nil {
			return
		}

		tronTo := to
		if tronTo[:2] == "0x" {
			tronTo, _ = tron.HexToTronAddress(to)
		}
		tronToken := token
		if token[:2] == "0x" && token != eth.EthAddress {
			tronToken, _ = tron.HexToTronAddress(tronToken)
		}

		var tx *api.TransactionExtention
		if token != eth.EthAddress {
			tx, err = conn.TRC20Send(TronFromAddress, tronTo, token, amount, 15000000)
			if err != nil {
				return
			}
		} else {
			tx, err = conn.Transfer(TronFromAddress, tronTo, amount.Int64())
			if err != nil {
				return
			}
		}
		signedTx, errr := tron.SignTx(priKey, tx.Transaction)
		if errr != nil {
			err = errr
			return
		}
		_, err = conn.Broadcast(signedTx)
		if err != nil {
			return
		}
		hash = common2.BytesToHexString(tx.GetTxid())
		return
	}
	client, err := ethclient.Dial(chain.Data.Rpc)
	if err != nil {
		return
	}

	privateKey, err := crypto.HexToECDSA(priKey)
	if err != nil {
		return
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err = client.NonceAt(context.Background(), fromAddress, nil)
	if err != nil {
		return
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chain.Data.ChainID))
	if err != nil {
		return
	}
	if token == eth.EthAddress {
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
		signedTx, err = types.SignTx(tx, types.NewEIP155Signer(big.NewInt(chain.Data.ChainID)), privateKey)
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
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = GasLimit   // in units
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
