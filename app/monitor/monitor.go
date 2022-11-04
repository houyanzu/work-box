package monitor

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/houyanzu/work-box/database/models/chainrecord"
	"github.com/houyanzu/work-box/database/models/chains"
	"math/big"
	"strings"
)

type EventLog struct {
	ChainDBID   uint
	logs        []types.Log
	netLastNum  uint64
	endBlockNum uint64
	contract    string
}

var initBlock = make(map[uint]map[string]uint64)

func InitBlockNum(chainDBID uint, contract string, blockNum uint64) {
	contract = strings.ToLower(contract)
	if initBlock[chainDBID] == nil {
		initBlock[chainDBID] = make(map[string]uint64)
	}
	initBlock[chainDBID][contract] = blockNum
}

func Monitor(chainDBID uint, contract string, blockDiff uint64) (res EventLog, err error) {
	contract = strings.ToLower(contract)
	chain := chains.New(nil).InitByID(chainDBID)
	if !chain.Exists() {
		err = errors.New("chain not found")
		return
	}
	client, err := ethclient.Dial(chain.Data.Rpc)
	if err != nil {
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
			record.Add()
		} else {
			panic("未初始化")
		}
	}
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return
	}
	netLastNum := header.Number.Uint64()
	endBlockNum := lastBlockNum + blockDiff

	contractAddress := common.HexToAddress(contract)
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(lastBlockNum + 1)),
		ToBlock:   big.NewInt(int64(endBlockNum)),
		Addresses: []common.Address{
			contractAddress,
		},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		return
	}
	res.logs = logs
	res.netLastNum = netLastNum
	res.endBlockNum = endBlockNum
	res.contract = contract
	res.ChainDBID = chainDBID
	return
}

func (e EventLog) Foreach(f func(index int, log types.Log, chainRecordId uint)) {
	have := false
	for k, v := range e.logs {
		blockNum := v.BlockNumber
		hash := v.TxHash.Hex()
		record := chainrecord.New(nil)
		record.Data.Contract = e.contract
		record.Data.BlockNum = blockNum
		record.Data.EventId = v.Topics[0].Hex()
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
