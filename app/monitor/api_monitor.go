package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/houyanzu/work-box/database/models/chainrecord"
	"github.com/houyanzu/work-box/database/models/chains"
	"github.com/houyanzu/work-box/lib/httptool"
	"github.com/houyanzu/work-box/tool/eth"
	"strings"
	"time"
)

type apiLogRes struct {
	Result []types.Log
}

func ApiMonitor(chainDBID uint, contract string, blockDiff uint64, eventID string) (res EventLog, err error) {
	contract = strings.ToLower(contract)

	chain := chains.New(nil).InitByID(chainDBID)
	if !chain.Exists() {
		err = errors.New("chain not found")
		return
	}

	lastBlockNum := chainrecord.GetLastBlockNum(chainDBID, contract)
	if lastBlockNum == 0 {
		var ok bool
		if lastBlockNum, ok = initBlock[chainDBID][contract]; ok {
			record := chainrecord.New(nil)
			record.Data.Contract = contract
			record.Data.BlockNum = lastBlockNum
			record.Data.EventId = ""
			record.Data.Hash = ""
			record.Data.ChainDbId = chainDBID
			record.Add()
		} else {
			panic("未初始化")
		}
	}
	netLastNum, err := eth.GetApiLastBlockNum(chainDBID)
	if err != nil {
		return
	}
	if netLastNum == 0 {
		err = errors.New("zero")
		return
	}
	endBlockNum := lastBlockNum + blockDiff

	url := chain.Data.ApiHost +
		"?module=logs&action=getLogs" +
		"&fromBlock=" + fmt.Sprintf("%d", lastBlockNum+1) +
		"&toBlock=" + fmt.Sprintf("%d", endBlockNum) +
		"&address=" + contract +
		"&apikey=" + chain.Data.ApiKey
	if eventID != "" {
		url += "&topic0=" + eventID
	}
	resp, code, err := httptool.Get(url, 20*time.Second)
	if err != nil {
		return
	}
	if code != 200 {
		err = errors.New(string(resp))
		return
	}

	str := strings.ReplaceAll(string(resp), `"logIndex":"0x"`, `"logIndex":"0x0"`)
	str = strings.ReplaceAll(str, `"transactionIndex":"0x"`, `"transactionIndex":"0x0"`)
	var logRes apiLogRes
	err = json.Unmarshal([]byte(str), &logRes)
	if err != nil {
		fmt.Println(str)
		return
	}
	res.logs = logRes.Result
	res.netLastNum = netLastNum
	res.endBlockNum = endBlockNum
	res.contract = contract
	res.ChainDBID = chainDBID
	return
}
