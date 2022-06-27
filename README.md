## work-box

```shell
go get github.com/houyanzu/work-box
```
 
### Init config

```go
package myconfig

import (
	"encoding/json"
	"github.com/houyanzu/work-box/config"
	"io/ioutil"
)

type MyConfig struct {
	config.Config
	Common commonConfig `json:"common"`
}

type commonConfig struct {
	Addr string `json:"addr"`
}

var myConfig *MyConfig

func ParseConfig(configName string) (err error) {
	dat, err := ioutil.ReadFile(configName)
	if err != nil {
		return err
	}

	err = json.Unmarshal(dat, &myConfig)
	if err != nil {
		return err
	}

	config.ParseConfig(&myConfig.Config)

	return nil
}

func GetConfig() *MyConfig {
	return myConfig
}
```

### monitor

```go
package main

import (
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/houyanzu/work-box/app/loop"
	"github.com/houyanzu/work-box/app/monitor"
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/contract/standardcoin"
	"github.com/houyanzu/work-box/lib/mylog"
	"sync"
	"test/myconfig"
)

func main() {
	var configName string
	var cycle int
	var startBlock uint64
	flag.Uint64Var(&startBlock, "startBlock", 0, "")
	flag.StringVar(&configName, "config", "", "Config json file name")
	flag.IntVar(&cycle, "cycle", 60, "")
	flag.Parse()

	err := myconfig.ParseConfig(configName)
	if err != nil {
		panic(err)
	}

	err = mylog.Init("monitor.log")
	if err != nil {
		panic(err)
	}
	err = database.InitMysql()
	if err != nil {
		panic(err)
	}

	conf := myconfig.GetConfig()
	monitor.InitBlockNum(conf.Common.Addr, startBlock)

	loop.Loop(cycle, mo)
}

func mo(wg *sync.WaitGroup) {
	defer wg.Done()
	conf := myconfig.GetConfig()
	logs, err := monitor.Monitor(conf.Common.Addr, 5000)
	if err != nil {
		mylog.Write(err)
		return
	}

	stc, err := standardcoin.NewStandardcoin(common.HexToAddress(""), nil)
	if err != nil {
		panic(err)
	}
	logs.Foreach(func(index int, log types.Log) {
		transferLog, err := stc.ParseTransfer(log)
		if err == nil {
			fmt.Printf("%+v", transferLog)
			return
		}

		approvalLog, err := stc.ParseApproval(log)
		if err == nil {
			fmt.Printf("%+v", approvalLog)
			return
		}
	})
}

```

### transfer

```go
package main

import (
	"flag"
	"fmt"
	"github.com/houyanzu/work-box/app/loop"
	"github.com/houyanzu/work-box/app/transfer"
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/mylog"
	"github.com/howeyc/gopass"
	"os"
	"sync"
	"test/lib/aes"
	"test/myconfig"
)

func main() {
	var configName string
	var keyFile string
	flag.StringVar(&keyFile, "keyFile", "", "keyFile")
	flag.StringVar(&configName, "config", "", "Config json file name")
	flag.Parse()

	err := myconfig.ParseConfig(configName)
	if err != nil {
		panic(err)
	}

	err = database.InitMysql()
	if err != nil {
		panic(err)
	}

	fmt.Println("enter password:")
	passByte, err := gopass.GetPasswdMasked()
	if err != nil {
		panic(err)
	}

	yottaPriKeyCt, err := os.ReadFile(keyFile)
	if err != nil {
		panic(err)
	}
	yottaPri := aes.NewFromBytes(yottaPriKeyCt)
	err = transfer.InitTrans(yottaPri, passByte)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = mylog.Init("transfer.log")

	loop.Loop(60, trans)
}

func trans(wg *sync.WaitGroup) {
	defer wg.Done()
	err := transfer.Transfer(20)
	if err != nil {
		mylog.Write(err)
	}
}

```