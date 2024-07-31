package balancewarn

import (
	"fmt"
	"github.com/houyanzu/work-box/database/models/balancewarning"
	"github.com/houyanzu/work-box/tool/dingtalk"
	"github.com/houyanzu/work-box/tool/eth"
	"github.com/shopspring/decimal"
)

func RunOnce() (err error) {
	warns := balancewarning.New(nil).GetAllOpen()
	fmt.Println(warns.List)

	warnString := "大王，不好啦！有情况：\n"
	warns.Foreach(func(key int, bw *balancewarning.Model) (b bool) {
		balance := decimal.Zero
		if bw.Data.Token == "0" {
			balance, err = eth.BalanceAt(bw.Data.ChainDBID, bw.Data.Address)
			if err != nil {
				return
			}
		} else {
			balance, err = eth.BalanceOf(bw.Data.ChainDBID, bw.Data.Token, bw.Data.Address)
			if err != nil {
				return
			}
		}
		fmt.Println(balance, bw.Data.WarningBalance)
		if balance.LessThan(bw.Data.WarningBalance) {
			warnString += fmt.Sprintf("地址：%s，余额：%s，低于预警值：%s，备注：%s\n", bw.Data.Address, balance.Div(decimal.New(1, bw.Data.Decimals)).String(), bw.Data.WarningBalance.Div(decimal.New(1, bw.Data.Decimals)).String(), bw.Data.Remark)
		}
		return
	})

	fmt.Println(warnString)
	if warnString != "大王，不好啦！有情况：\n" {
		err = dingtalk.PushMsg(warnString)
	}
	return
}
