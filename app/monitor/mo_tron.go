package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/houyanzu/work-box/database/models/chainrecord"
	"github.com/houyanzu/work-box/database/models/chains"
	"github.com/houyanzu/work-box/lib/httptool"
	"strings"
	"time"
)

type TronEventLog struct {
	ChainDBID   uint
	logs        []TronEventResData
	netLastNum  uint64
	endBlockNum uint64
	contract    string
}

type tronEventRes struct {
	Data    []TronEventResData `json:"data"`
	Success bool               `json:"success"`
	Meta    struct {
		At          int64  `json:"at"`
		Fingerprint string `json:"fingerprint"`
		Links       struct {
			Next string `json:"next"`
		} `json:"links"`
		PageSize int `json:"page_size"`
	} `json:"meta"`
}

type TronEventResData struct {
	BlockNumber           uint64 `json:"block_number"`
	BlockTimestamp        uint64 `json:"block_timestamp"`
	CallerContractAddress string `json:"caller_contract_address"`
	ContractAddress       string `json:"contract_address"`
	EventIndex            int    `json:"event_index"`
	EventName             string `json:"event_name"`
	Result                struct {
		Field1 string `json:"0"`
		Field2 string `json:"1"`
		Field3 string `json:"2"`
		From   string `json:"from"`
		To     string `json:"to"`
		Value  string `json:"value"`
	} `json:"result"`
	ResultType struct {
		From  string `json:"from"`
		To    string `json:"to"`
		Value string `json:"value"`
	} `json:"result_type"`
	Event         string `json:"event"`
	TransactionId string `json:"transaction_id"`
}

func MonitorTron(chainDBID uint, contract string, endBlockTimestamp uint64, eventID string) (res TronEventLog, err error) {
	chain := chains.New(nil).InitByID(chainDBID)
	if !chain.Exists() {
		err = errors.New("chain not found")
		return
	}

	lastBlockTimestamp := chainrecord.GetLastBlockNum(chainDBID, contract)
	if lastBlockTimestamp == 0 {
		var ok bool
		contractLower := strings.ToLower(contract)
		if lastBlockTimestamp, ok = initBlock[chainDBID][contractLower]; ok {
			record := chainrecord.New(nil)
			record.Data.Contract = contract
			record.Data.BlockNum = lastBlockTimestamp
			record.Data.EventId = ""
			record.Data.Hash = ""
			record.Data.ChainDbId = chainDBID
			record.Add()
		} else {
			panic("未初始化")
		}
	}

	if endBlockTimestamp == 0 {
		endBlockTimestamp = lastBlockTimestamp + 120000
	}

	nowTimestamp := time.Now().UnixMilli()
	if uint64(nowTimestamp) < endBlockTimestamp {
		endBlockTimestamp = uint64(nowTimestamp)
	}

	for lastBlockTimestamp < endBlockTimestamp {
		end := lastBlockTimestamp + 30000
		if end > endBlockTimestamp {
			end = endBlockTimestamp
		}
		fingerprint := ""
		for {
			url := chain.Data.ApiHost + "contracts/" + contract + "/events" +
				"?only_confirmed=true&order_by=block_timestamp,asc&limit=200" +
				"&min_block_timestamp=" + fmt.Sprintf("%d", lastBlockTimestamp+1) +
				"&max_block_timestamp=" + fmt.Sprintf("%d", end)
			if eventID != "" {
				url += "&event_name=" + eventID
			}
			url += "&fingerprint=" + fingerprint
			httptool.Header = make(map[string]string)
			httptool.Header["TRON_PRO_API_KEY"] = chain.Data.ApiKey
			resp, code, errr := httptool.Get(url, 20*time.Second)
			if errr != nil {
				err = errr
				return
			}
			if code != 200 {
				err = errors.New(string(resp))
				return
			}
			var logRes tronEventRes
			err = json.Unmarshal(resp, &logRes)
			if err != nil {
				return
			}
			if !logRes.Success {
				err = errors.New(string(resp))
				return
			}

			res.logs = append(res.logs, logRes.Data...)

			fingerprint = logRes.Meta.Fingerprint
			if fingerprint == "" {
				break
			}
		}
		lastBlockTimestamp = end
	}
	res.contract = contract
	res.ChainDBID = chainDBID
	res.endBlockNum = endBlockTimestamp
	res.netLastNum = uint64(nowTimestamp)

	return
}

func (e TronEventLog) Foreach(f func(index int, log TronEventResData, chainRecordId uint)) {
	have := false
	for k, v := range e.logs {
		blockNum := v.BlockTimestamp
		hash := v.TransactionId
		record := chainrecord.New(nil)
		record.Data.Contract = e.contract
		record.Data.BlockNum = blockNum
		record.Data.EventId = v.EventName
		record.Data.Hash = hash
		record.Data.ChainDbId = e.ChainDBID
		record.Add()
		f(k, v, record.Data.ID)
		have = true
	}
	if !have {
		if e.endBlockNum <= e.netLastNum {
			record := chainrecord.New(nil)
			record.Data.Contract = e.contract
			record.Data.BlockNum = e.endBlockNum
			record.Data.ChainDbId = e.ChainDBID
			record.Add()
		}
	}
}
