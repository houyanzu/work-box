package feed

import (
	"github.com/houyanzu/work-box/database/models/chains"
	"github.com/houyanzu/work-box/database/models/transferdetails"
	"github.com/houyanzu/work-box/database/models/userkeys"
	"github.com/houyanzu/work-box/database/models/userkeysbalance"
	"github.com/houyanzu/work-box/tool/eth"
	"sync"
)

const TransferModule = "FEED_USER_KEY"

var ChainDBID = uint(1)

func Feed(wg *sync.WaitGroup) {
	defer wg.Done()

	chain := chains.New(nil).InitByID(ChainDBID)
	feedingList := userkeysbalance.New(nil).InitFeedingList(ChainDBID)
	if !feedingList.ListEmpty() {
		feedingList.Foreach(func(index int, feeding *userkeysbalance.Model) {
			transferDetail := transferdetails.New(nil).InitByID(feeding.Data.ModuleID)
			if transferDetail.Data.Status == 2 {
				feeding.SetFeedFinish()
			}
		})
	}

	waitList := userkeysbalance.New(nil).InitWaitingList()
	if waitList.ListEmpty() {
		return
	}

	waitList.Foreach(func(index int, userKeyBa *userkeysbalance.Model) {
		userKey := userkeys.New(nil).InitById(userKeyBa.Data.KeyID)
		transferDetail := transferdetails.New(nil)
		transferDetail.Data.Module = TransferModule
		transferDetail.Data.Token = eth.EthAddress
		transferDetail.Data.To = userKey.Data.Address
		transferDetail.Data.Amount = chain.Data.UserKeyFeedAmount
		transferDetail.Data.ChainDbId = ChainDBID
		transferDetail.Add()

		userKeyBa.SetFeeding(transferDetail.Data.ID)
	})
}
