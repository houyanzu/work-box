package feed

import (
	"github.com/houyanzu/work-box/config"
	"github.com/houyanzu/work-box/database/models/transferdetails"
	"github.com/houyanzu/work-box/database/models/userkeys"
	"github.com/houyanzu/work-box/tool/eth"
	"sync"
)

const TransferModule = "FEED_USER_KEY"

func Feed(wg *sync.WaitGroup) {
	defer wg.Done()

	feedingList := userkeys.New(nil).InitFeedingList()
	if !feedingList.ListEmpty() {
		feedingList.Foreach(func(index int, feeding *userkeys.Model) {
			transferDetail := transferdetails.New(nil).InitByID(feeding.Data.TransferDetailID)
			if transferDetail.Data.Status == 2 {
				feeding.SetFeedFinish()
			}
		})
	}

	waitList := userkeys.New(nil).InitWaitingList()
	if waitList.ListEmpty() {
		return
	}

	conf := config.GetConfig()

	waitList.Foreach(func(index int, userKey *userkeys.Model) {
		transferDetail := transferdetails.New(nil)
		transferDetail.Data.Module = TransferModule
		transferDetail.Data.Token = eth.EthAddress
		transferDetail.Data.To = userKey.Data.Address
		transferDetail.Data.Amount = conf.Extra.UserKeyFeedAmount
		transferDetail.Add()

		userKey.SetFeeding(transferDetail.Data.ID)
	})
}
