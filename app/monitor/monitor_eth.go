package monitor

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/houyanzu/work-box/database/models/chainrecord"
	"github.com/houyanzu/work-box/database/models/chains"
	"github.com/houyanzu/work-box/tool/eth"
	"math/big"
)

type EthLog struct {
	ChainDBID   uint
	logs        []transferLog
	netLastNum  uint64
	endBlockNum uint64
	contract    string
}

type transferLog struct {
	TxHash      common.Hash
	BlockNumber uint64
	From        common.Address
	To          common.Address
	Amount      *big.Int
}

func MonitorEth(chainDBID uint, blockDiff uint64) (res EthLog, err error) {
	contract := eth.EthAddress
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
	startBlockNum := lastBlockNum + 1
	endBlockNum := lastBlockNum + blockDiff

	logs := make([]transferLog, 0)
	for i := startBlockNum; i <= endBlockNum; i++ {
		var block *types.Block
		block, err = client.BlockByNumber(context.Background(), big.NewInt(int64(i)))
		if err != nil {
			block, err = client.BlockByNumber(context.Background(), big.NewInt(int64(i)))
			if err != nil {
				continue
			}
		}
		for _, tx := range block.Transactions() {
			msg, _ := tx.AsMessage(types.NewEIP155Signer(big.NewInt(chain.Data.ChainID)), nil)
			logs = append(logs, transferLog{tx.Hash(), i, msg.From(), *tx.To(), tx.Value()})
		}
	}

	res.logs = logs
	res.netLastNum = netLastNum
	res.endBlockNum = endBlockNum
	res.contract = contract
	res.ChainDBID = chainDBID
	return
}

func (e EthLog) Foreach(f func(index int, log transferLog)) {
	for k, v := range e.logs {
		blockNum := v.BlockNumber
		hash := v.TxHash.Hex()
		record := chainrecord.New(nil)
		record.Data.Contract = e.contract
		record.Data.BlockNum = blockNum
		record.Data.EventId = ""
		record.Data.Hash = hash
		record.Data.ChainDbId = e.ChainDBID
		record.Add()
		f(k, v)
	}
	if e.endBlockNum <= e.netLastNum {
		record := chainrecord.New(nil)
		record.Data.Contract = e.contract
		record.Data.BlockNum = e.endBlockNum
		record.Data.ChainDbId = e.ChainDBID
		record.Add()
	}
}
