package transfer

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/houyanzu/work-box/config"
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
	"github.com/houyanzu/work-box/tool/eth"
	"math/big"
)

var privateKeyStr string
var FromAddress string
var GasLimit = uint64(150000)

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
	return
}

func Transfer(limit int, module string) (err error) {
	conf := config.GetConfig()
	pending := transferrecords.New(nil).InitPending(FromAddress, module)
	if pending.Exists() {
		var status uint64
		status, err = eth.GetTxStatus(pending.Data.Hash)
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

	waitingList := transferdetails.New(nil).InitWaitingList(limit, module)
	length := len(waitingList.List)
	if length == 0 {
		return
	}

	tokens, tos := make([]common.Address, length), make([]common.Address, length)
	ids := make([]uint, length)
	amounts := make([]*big.Int, length)
	totalValue := big.NewInt(0)
	waitingList.Foreach(func(index int, m *transferdetails.Model) {
		tokens[index] = common.HexToAddress(m.Data.Token)
		tos[index] = common.HexToAddress(m.Data.To)
		amount := m.Data.Amount.BigInt()
		amounts[index] = amount
		ids[index] = m.Data.ID
		if m.Data.Token == "0x0000000000000000000000000000000000000000" {
			totalValue.Add(totalValue, amount)
		}
	})
	client, err := ethclient.Dial(conf.Eth.Host)
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

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(conf.Eth.ChainId))
	if err != nil {
		return
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = totalValue  // in wei
	auth.GasLimit = GasLimit // in units
	auth.GasPrice = gasPrice

	multiCon := common.HexToAddress(conf.Eth.MultiTransferContract)
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
	tr.Data.From = FromAddress
	tr.Data.Hash = hash
	tr.Data.Nonce = nonce
	tr.Data.Module = module
	tr.Add()

	transferdetails.New(nil).SetExec(ids, tr.Data.ID)
	return
}

func LockTransfer(module string) (err error) {
	limit := 1
	conf := config.GetConfig()
	pending := transferrecords.New(nil).InitPending(FromAddress, module)
	if pending.Data.ID > 0 {
		var status uint64
		status, err = eth.GetTxStatus(pending.Data.Hash)
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

	waitingList := locktransferdetails.New(nil).InitWaitingList(limit, module)
	length := len(waitingList.List)
	if length == 0 {
		return
	}

	waiting := locktransferdetails.New(nil).InitByData(waitingList.List[0])

	client, err := ethclient.Dial(conf.Eth.Host)
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

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(conf.Eth.ChainId))
	if err != nil {
		return
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = GasLimit   // in units
	auth.GasPrice = gasPrice

	ltCon := common.HexToAddress(conf.Eth.LockTransferContract)
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
	tr.Data.Type = 2
	tr.Data.From = FromAddress
	tr.Data.Hash = hash
	tr.Data.Nonce = nonce
	tr.Add()

	locktransferdetails.New(nil).SetExec([]uint{waiting.Data.ID}, tr.Data.ID)
	return
}

func SingleTransfer(token string, to string, amount *big.Int, priKey string) (hash string, nonce uint64, err error) {
	conf := config.GetConfig()
	client, err := ethclient.Dial(conf.Eth.Host)
	if err != nil {
		return
	}

	if priKey == "" {
		priKey = privateKeyStr
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

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(conf.Eth.ChainId))
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
		signedTx, err = types.SignTx(tx, types.NewEIP155Signer(big.NewInt(conf.Eth.ChainId)), privateKey)
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
