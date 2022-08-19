package monitor

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/houyanzu/work-box/config"
)

var (
	client  *ethclient.Client
	sub     ethereum.Subscription
	client2 *ethclient.Client
	sub2    ethereum.Subscription
	logs    *chan types.Log
	current int = 1
)

func initClient(contract string) (err error) {
	conf := config.GetConfig()
	if current == 1 {
		client, err = ethclient.Dial(conf.Eth.WssHost)
		if err != nil {
			fmt.Println(111, err)
			return
		}
		contractAddress := common.HexToAddress(contract)
		query := ethereum.FilterQuery{
			Addresses: []common.Address{contractAddress},
		}

		l := make(chan types.Log)
		sub, err = client.SubscribeFilterLogs(context.Background(), query, l)
		if err != nil {
			return
		}
		logs = &l
		if sub2 != nil {
			sub.Unsubscribe()
		}
		if client2 != nil {
			client.Close()
		}
		current = 0
		return
	}
	client2, err = ethclient.Dial(conf.Eth.WssHost)
	if err != nil {
		return
	}
	contractAddress := common.HexToAddress(contract)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	l := make(chan types.Log)
	sub2, err = client.SubscribeFilterLogs(context.Background(), query, l)
	if err != nil {
		return
	}
	logs = &l
	if sub != nil {
		sub.Unsubscribe()
	}
	if client != nil {
		client.Close()
	}
	current = 1
	return
}

func Listen(contract string) (logss *chan types.Log, err error) {
	err = initClient(contract)
	if err != nil {
		return
	}
	logss = logs

	//go func() {
	//	for {
	//		time.Sleep(5 * time.Minute)
	//		_ = initClient(contract)
	//	}
	//}()
	return
}
