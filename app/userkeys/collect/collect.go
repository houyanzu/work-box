package collect

import (
	"errors"
	"github.com/houyanzu/work-box/app/transfer"
	"github.com/houyanzu/work-box/database/models/chains"
	"github.com/houyanzu/work-box/database/models/keys"
	"github.com/houyanzu/work-box/database/models/tokens"
	"github.com/houyanzu/work-box/database/models/ukcollectrecord"
	"github.com/houyanzu/work-box/database/models/userkeys"
	"github.com/houyanzu/work-box/database/models/userkeysbalance"
	crypto2 "github.com/houyanzu/work-box/lib/crypto"
	"github.com/houyanzu/work-box/tool/eth"
)

var CollectGasLimit = uint64(150000)

func Collect(chainDBID uint, password []byte, ukbID, toKeyID uint, de crypto2.Decoder) (err error) {
	ukb := userkeysbalance.New(nil).InitById(ukbID)
	if !ukb.Exists() {
		err = errors.New("balance not exists")
		return
	}

	uk := userkeys.New(nil).InitById(ukb.Data.KeyID)
	if ukb.Data.CollectStatus == 1 {
		collectRecord := ukcollectrecord.New(nil).InitPendingByKeyID(chainDBID, uk.Data.ID)
		if collectRecord.Exists() {
			status, err := eth.GetTxStatus(chainDBID, collectRecord.Data.Hash)
			if err != nil {
				return err
			}
			if status == 1 {
				collectRecord.SetSuccess()
			} else {
				collectRecord.SetFail()
			}
			ukb.SetCollectFinish()
			ukb.UpdateBalance()
			return nil
		}
	}
	balance, err := eth.BalanceAt(chainDBID, uk.Data.Address)
	if err != nil {
		return
	}
	chain := chains.New(nil).InitByID(chainDBID)
	if !chain.Exists() {
		err = errors.New("chain not exists")
		return
	}
	if balance.LessThan(chain.Data.UserKeyFeedAmount) {
		ukb.SetWaitFeed()
		return
	}
	if ukb.Data.Status != 0 {
		return
	}

	toKey := keys.New(nil).InitByID(toKeyID)

	transfer.GasLimit = CollectGasLimit
	token := tokens.New(nil).InitById(ukb.Data.TokenID)
	amount := ukb.Data.Balance
	hash, nonce, err := transfer.SingleTransfer(chainDBID, token.Data.Contract, toKey.Data.Address, amount.BigInt(), uk.GetPriKey(password, de))
	if err != nil {
		return
	}

	cr := ukcollectrecord.New(nil)
	cr.Data.KeyID = uk.Data.ID
	cr.Data.Hash = hash
	cr.Data.Status = 1
	cr.Data.Nonce = nonce
	cr.Data.Amount = amount
	cr.Data.BalanceID = ukb.Data.ID
	cr.Data.ChainDbId = chainDBID
	cr.Add()

	ukb.SetCollecting(cr.Data.ID)

	return
}
