package collect

import (
	"errors"
	"github.com/houyanzu/work-box/app/transfer"
	"github.com/houyanzu/work-box/database/models/collecttokens"
	"github.com/houyanzu/work-box/database/models/keys"
	"github.com/houyanzu/work-box/database/models/tokens"
	"github.com/houyanzu/work-box/database/models/ukcollectrecord"
	"github.com/houyanzu/work-box/database/models/userkeys"
	"github.com/houyanzu/work-box/database/models/userkeysbalance"
	crypto2 "github.com/houyanzu/work-box/lib/crypto"
	"github.com/houyanzu/work-box/lib/mytime"
	"github.com/houyanzu/work-box/tool/eth"
	"time"
)

var CollectGasLimit = uint64(150000)

func Collect(chainDBID uint, password []byte, ukbID, toKeyID uint, de crypto2.Decoder) (err error) {
	ukb := userkeysbalance.New(nil).InitById(ukbID)
	if !ukb.Exists() {
		err = errors.New("balance not exists")
		return
	}

	uk := userkeys.New(nil).InitById(ukb.Data.KeyID)

	pendingUkb := userkeysbalance.New(nil).InitCollectingByKeyID(uk.Data.ID)

	if pendingUkb.Exists() {
		collectRecord := ukcollectrecord.New(nil).InitPendingByKeyID(chainDBID, uk.Data.ID)
		if collectRecord.Exists() {
			status, err := eth.GetTxStatus(chainDBID, collectRecord.Data.Hash)
			if err != nil {
				exTime := collectRecord.Data.CreateTime.Add(30 * time.Minute)
				now := mytime.NewFromNow()
				if exTime.Before(now) {
					collectRecord.SetFail()
					pendingUkb.SetCollectFinish()
					pendingUkb.UpdateBalance()
				}
				return nil
			}
			if status == 1 {
				collectRecord.SetSuccess()
			} else {
				collectRecord.SetFail()
			}
			pendingUkb.SetCollectFinish()
			pendingUkb.UpdateBalance()
			return nil
		}
	}
	if ukb.Data.Status != 0 {
		return
	}
	balance, err := eth.BalanceAt(chainDBID, uk.Data.Address)
	if err != nil {
		return
	}

	ct := collecttokens.New(nil).InitByTokenID(ukb.Data.TokenID)
	if balance.LessThan(ct.Data.FeedAmount) {
		ukb.SetWaitFeed()
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
