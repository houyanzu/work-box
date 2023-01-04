package transfer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	common2 "github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/houyanzu/work-box/lib/contract/standardcoin"
	"github.com/houyanzu/work-box/lib/tron"
	"google.golang.org/grpc"
	"math/big"
	"testing"
	"time"
)

func TestTransfer(t *testing.T) {
	hash, err := SingleTransfer2("0x0000000000000000000000000000000000000000", "0x927767f07bA44CDcC560DB5DE3c915159d309d1A", big.NewInt(10000000000))
	if err != nil {
		panic(err)
	}
	fmt.Println(hash)
}

func SingleTransfer2(token string, to string, amount *big.Int) (hash string, err error) {
	client, err := ethclient.Dial("https://bsc-dataseed1.defibit.io/")
	if err != nil {
		return
	}

	privateKey, err := crypto.HexToECDSA("6efef37df82d19ccf2f151ccb1a6b9db5909b19f7ac0d6072e6c0d26c75ea61c")
	if err != nil {
		return
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.NonceAt(context.Background(), fromAddress, nil)
	if err != nil {
		return
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(56))
	if err != nil {
		return
	}
	if token == "0x0000000000000000000000000000000000000000" {
		//tx := types.NewTransaction(nonce, common.HexToAddress(to), amount, 21000, gasPrice, nil)
		toAddress := common.HexToAddress(to)
		baseTx := &types.LegacyTx{
			To:       &toAddress,
			Nonce:    nonce,
			Value:    amount,
			Gas:      21000,
			GasPrice: gasPrice,
			Data:     nil,
		}
		tx := types.NewTx(baseTx)
		var signedTx *types.Transaction
		signedTx, err = types.SignTx(tx, types.NewEIP155Signer(big.NewInt(56)), privateKey)
		if err != nil {
			return
		}

		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			return
		}
		hash = signedTx.Hash().Hex()
		return
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(150000) // in units
	auth.GasPrice = gasPrice

	tokenCon := common.HexToAddress(token)
	tokenInstance, err := standardcoin.NewStandardcoin(tokenCon, client)
	if err != nil {
		return
	}
	tx, err := tokenInstance.Transfer(auth, common.HexToAddress(to), amount)
	if err != nil {
		return
	}
	hash = tx.Hash().Hex()
	return
}

var (
	conn                  *client.GrpcClient
	apiKey                = "1dd9e753-141c-4389-9d50-30b9eefcb6d2"
	tronAddress           = "grpc.trongrid.io:50051"
	accountAddress        = "TPpw7soPWEDQWXPCGUMagYPryaWrYR5b3b"
	accountAddressWitness = "TGj1Ej1qRzL9feLTLhjwgxXF4Ct6GTWg2U"
)

func TestTron(t *testing.T) {
	//type con struct {
	//	Parameter struct {
	//		Value struct {
	//			Amount       int    `json:"amount" gorm:"column:amount"`
	//			OwnerAddress string `json:"owner_address" gorm:"column:owner_address"`
	//			ToAddress    string `json:"to_address" gorm:"column:to_address"`
	//		} `json:"value" gorm:"column:value"`
	//		TypeUrl string `json:"type_url" gorm:"column:type_url"`
	//	} `json:"parameter" gorm:"column:parameter"`
	//	Type string `json:"type" gorm:"column:type"`
	//}
	//type rawData struct {
	//	Contract      []con  `json:"contract" gorm:"column:contract"`
	//	RefBlockBytes string `json:"ref_block_bytes" gorm:"column:ref_block_bytes"`
	//	RefBlockHash  string `json:"ref_block_hash" gorm:"column:ref_block_hash"`
	//	Expiration    int64  `json:"expiration" gorm:"column:expiration"`
	//	Timestamp     int64  `json:"timestamp" gorm:"column:timestamp"`
	//}

	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithInsecure())

	conn = client.NewGrpcClient(tronAddress)
	conn.SetTimeout(5 * time.Second)

	if err := conn.Start(opts...); err != nil {
		_ = fmt.Errorf("Error connecting GRPC Client: %v", err)
	}

	err := conn.SetAPIKey(apiKey)
	if err != nil {
		panic(err)
	}

	tx, err := conn.Transfer("TSjaahyvbDjFeGpESg3S99WsNEroCnrHGw", "TTp3xVVgsg4rfSKB5xLfcrp85n5HKvKcny", 1)
	if err != nil {
		panic(err)
	}
	signedTx, errr := tron.SignTx("1cb80aad8b187aec43796cf0a382ac9e75c8703866a4bfd0d5134b387008f2d5", tx.Transaction)
	if errr != nil {
		err = errr
		return
	}
	//res, err := conn.Broadcast(signedTx)
	//if err != nil {
	//	panic(err)
	//}
	//contr := con{
	//	Parameter: struct {
	//		Value struct {
	//			Amount       int    `json:"amount" gorm:"column:amount"`
	//			OwnerAddress string `json:"owner_address" gorm:"column:owner_address"`
	//			ToAddress    string `json:"to_address" gorm:"column:to_address"`
	//		} `json:"value" gorm:"column:value"`
	//		TypeUrl string `json:"type_url" gorm:"column:type_url"`
	//	}{},
	//	Type: "",
	//}
	//rawDatas := rawData{
	//	Contract:      nil,
	//	RefBlockBytes: "",
	//	RefBlockHash:  "",
	//	Expiration:    0,
	//	Timestamp:     0,
	//}
	fmt.Println(signedTx.String())
	//js, _ := json.Marshal(&signedTx)
	//httptool.PostJSON("https://api.trongrid.io/wallet/broadcasttransaction", signedTx)
	//fmt.Println(res)
}

func TestTronCon(t *testing.T) {
	//ks := keystore.ForPath("D:\\work\\gowork\\work-box\\app\\transfer")
	//fmt.Println("")

	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithInsecure())

	conn = client.NewGrpcClient(tronAddress)

	if err := conn.Start(opts...); err != nil {
		_ = fmt.Errorf("Error connecting GRPC Client: %v", err)
	}

	err := conn.SetAPIKey(apiKey)
	if err != nil {
		panic(err)
	}

	tx, err := conn.TRC20Send("TTp3xVVgsg4rfSKB5xLfcrp85n5HKvKcny", "TQs81QzrVaSnG4UjVQ6aJgZS3q6VKSRJS4", "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t", big.NewInt(100000), 15000000)
	if err != nil {
		panic(err)
	}
	//addr, err := address.Base58ToAddress("TSjaahyvbDjFeGpESg3S99WsNEroCnrHGw")
	//if err != nil {
	//	panic(err)
	//}
	//err = ks.Unlock(keystore.Account{Address: addr}, "0123456789123456")
	//if err != nil {
	//	panic(err)
	//}
	signedTx, err := tron.SignTx("", tx.Transaction)
	//signedTx, err := tron.SignTx("", tx.Transaction)
	if err != nil {
		panic(err)
	}
	_, err = conn.Broadcast(signedTx)
	if err != nil {
		panic(err)
	}
	fmt.Println(common2.BytesToHexString(tx.GetTxid()))
}

func TestConvert(t *testing.T) {
	fmt.Println(tron.HexToTronAddress("0xa1c17d154b123665ce05ad8e6a581d3499a5c49b"))
	//fmt.Println(tron.TronAddressToHex("TQiVamUvkCkSgAvb1irTYCD1AbeM4uVfVa"))
}

func TestEn(t *testing.T) {
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithInsecure())

	conn = client.NewGrpcClient(tronAddress)

	if err := conn.Start(opts...); err != nil {
		_ = fmt.Errorf("Error connecting GRPC Client: %v", err)
	}

	err := conn.SetAPIKey(apiKey)
	if err != nil {
		panic(err)
	}
	tx, err := conn.FreezeBalance("TSjaahyvbDjFeGpESg3S99WsNEroCnrHGw", "TTp3xVVgsg4rfSKB5xLfcrp85n5HKvKcny", core.ResourceCode_ENERGY, 50000000)
	if err != nil {
		panic(err)
	}
	//addr, err := address.Base58ToAddress("TSjaahyvbDjFeGpESg3S99WsNEroCnrHGw")
	//if err != nil {
	//	panic(err)
	//}
	//err = ks.Unlock(keystore.Account{Address: addr}, "0123456789123456")
	//if err != nil {
	//	panic(err)
	//}
	signedTx, err := tron.SignTx("", tx.Transaction)
	if err != nil {
		panic(err)
	}
	_, err = conn.Broadcast(signedTx)
	if err != nil {
		panic(err)
	}
}

func TestBa(t *testing.T) {
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithInsecure())

	conn = client.NewGrpcClient(tronAddress)

	if err := conn.Start(opts...); err != nil {
		_ = fmt.Errorf("Error connecting GRPC Client: %v", err)
	}

	err := conn.SetAPIKey(apiKey)
	if err != nil {
		panic(err)
	}

	balance, err := conn.TRC20ContractBalance("TSjaahyvbDjFeGpESg3S99WsNEroCnrHGw", "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t")
	if err != nil {
		panic(err)
	}

	account, err := conn.GetAccount("TSjaahyvbDjFeGpESg3S99WsNEroCnrHGw")
	if err != nil {
		panic(err)
	}

	fmt.Println("USDT:", balance)
	fmt.Println("TRX:", account.Balance)
	fmt.Println("frozen_TRX:", account.Frozen[0].FrozenBalance)
}

func TestTrx(t *testing.T) {
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithInsecure())

	conn = client.NewGrpcClient(tronAddress)

	if err := conn.Start(opts...); err != nil {
		_ = fmt.Errorf("Error connecting GRPC Client: %v", err)
	}

	err := conn.SetAPIKey(apiKey)
	if err != nil {
		panic(err)
	}

	//trx, err := conn.GetTransactionByID("")
	trx, err := conn.GetTransactionByID("0988b7f18bdeba4cab5ae4cd352800983154d08b6ac743690298c4f8a407c0a0")
	if err != nil {
		panic(err)
	}
	fmt.Println(trx.Ret[0].ContractRet)
}
