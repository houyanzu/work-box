package transfer

import (
	"errors"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	common2 "github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/houyanzu/work-box/database/models/chains"
	"github.com/houyanzu/work-box/database/models/locktransferdetails"
	"github.com/houyanzu/work-box/database/models/transferdetails"
	"github.com/houyanzu/work-box/database/models/transferrecords"
	"github.com/houyanzu/work-box/lib/mytime"
	"github.com/houyanzu/work-box/lib/tron"
	"github.com/houyanzu/work-box/tool/eth"
	"google.golang.org/grpc"
	"time"
)

func TronTransfer(chainDBID uint, limit int, module string) (err error) {
	limit = 1
	chain := chains.New(nil).InitByID(chainDBID)
	if !chain.Exists() {
		err = errors.New("chain not found")
		return
	}
	pending := transferrecords.New(nil).InitPending(chainDBID, TronFromAddress, module)
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

	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithInsecure())

	conn := client.NewGrpcClient(chain.Data.Rpc)

	if err = conn.Start(opts...); err != nil {
		return
	}

	err = conn.SetAPIKey(chain.Data.ApiKey)
	if err != nil {
		return
	}

	if CheckBalance {
		balance, errr := eth.BalanceOf(chainDBID, waitingList.List[0].Token, TronFromAddress)
		if errr != nil {
			err = errr
			return
		}
		if balance.LessThan(waitingList.List[0].Amount) {
			err = errors.New("insufficient balance: " + waitingList.List[0].Token)
			return
		}
	}

	tx, err := conn.TRC20Send(TronFromAddress, waitingList.List[0].To, waitingList.List[0].Token, waitingList.List[0].Amount.BigInt(), 15000000)
	if err != nil {
		return
	}
	if waitingList.List[0].Token == "" {
		tx, err = conn.Transfer(TronFromAddress, waitingList.List[0].To, waitingList.List[0].Amount.IntPart())
		if err != nil {
			panic(err)
		}
	}
	signedTx, err := tron.SignTx(privateKeyStr, tx.Transaction)
	if err != nil {
		return
	}
	_, err = conn.Broadcast(signedTx)
	if err != nil {
		return
	}

	tr := transferrecords.New(nil)
	tr.Data.Type = 1
	tr.Data.ChainDbId = chainDBID
	tr.Data.From = TronFromAddress
	tr.Data.Hash = common2.BytesToHexString(tx.GetTxid())
	tr.Data.Nonce = 0
	tr.Data.Module = module
	tr.Add()

	transferdetails.New(nil).SetExec([]uint{waitingList.List[0].ID}, tr.Data.ID)
	return
}
