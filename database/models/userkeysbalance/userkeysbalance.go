package userkeysbalance

import (
	"fmt"
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/database/models/tokens"
	"github.com/houyanzu/work-box/database/models/userkeys"
	"github.com/houyanzu/work-box/tool/eth"
	"github.com/shopspring/decimal"
)

type BoxUserKeysBalance struct {
	ID            uint            `json:"id" gorm:"column:id"`
	ChainDbID     uint            `json:"chain_db_id" gorm:"column:chain_db_id"`
	KeyID         uint            `json:"key_id" gorm:"column:key_id"`
	TokenID       uint            `json:"token_id" gorm:"column:token_id"`
	Balance       decimal.Decimal `json:"balance" gorm:"column:balance"`
	Status        int8            `json:"status" gorm:"column:status"`                 // 0-正常，1-待转eth，2-转eth中
	CollectStatus int8            `json:"collect_status" gorm:"column:collect_status"` // 0-正常，1-归集中
	ModuleID      uint            `json:"module_id" gorm:"column:module_id"`
}

func (m *BoxUserKeysBalance) TableName() string {
	return "box_user_keys_balance"
}

var haveTable = false

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_user_keys_balance` (\n\t`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`chain_db_id` int(11) UNSIGNED NOT NULL DEFAULT 1,\n\t`key_id` int(11) UNSIGNED NOT NULL,\n\t`token_id` int(11) UNSIGNED NOT NULL,\n\t`balance` decimal(32,0)  NOT NULL,\n\t`status` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0-正常，1-待转eth，2-转eth中',\n\t`collect_status` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0-正常，1-归集中',\n\t`module_id` int(11) UNSIGNED NOT NULL DEFAULT 0,\n\tPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci\nAUTO_INCREMENT=1\nROW_FORMAT=DYNAMIC\nAVG_ROW_LENGTH=0;"
	return db.Exec(sql).Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxUserKeysBalance
	List  []BoxUserKeysBalance
	Total int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxUserKeysBalance, 0)
	data := BoxUserKeysBalance{}
	if !haveTable {
		hasTable := ctx.Db.Migrator().HasTable(&data)
		if !hasTable {
			err := createTable()
			if err != nil {
				panic(err)
			}
		}
		haveTable = true
	}

	return &Model{ctx, data, list, 0}
}

func (m *Model) Add() {
	m.Db.Create(&m.Data)
}

func (m *Model) Exists() bool {
	return m.Data.ID > 0
}

func (m *Model) InitByKeyAndToken(keyID, tokenID uint) *Model {
	token := tokens.New(nil).InitById(tokenID)
	m.Db.Where("key_id = ? AND token_id = ?", keyID, tokenID).Take(&m.Data)
	if !m.Exists() {
		m.Data.KeyID = keyID
		m.Data.TokenID = tokenID
		m.Data.ChainDbID = token.Data.ChainDbID
	}
	return m
}

func (m *Model) UpdateBalance() {
	token := tokens.New(nil).InitById(m.Data.TokenID)
	userKey := userkeys.New(nil).InitById(m.Data.KeyID)
	m.Data.Balance, _ = eth.BalanceOf(m.Data.ChainDbID, token.Data.Contract, userKey.Data.Address)
	m.Db.Save(&m.Data)
}

func (m *Model) InitListByBalance(chainDBID, tokenID uint, balance decimal.Decimal) *Model {
	m.Db.Where("chain_db_id = ? AND token_id = ?", chainDBID, tokenID).Where(fmt.Sprintf("balance >= %s", balance.String())).Find(&m.List)
	return m
}

func (m *Model) InitById(id uint) *Model {
	m.Db.Take(&m.Data, id)
	return m
}

func (m *Model) InitByData(data BoxUserKeysBalance) *Model {
	m.Data = data
	return m
}

func (m *Model) Foreach(f func(index int, m *Model)) {
	for k, v := range m.List {
		mm := New(nil).InitByData(v)
		f(k, mm)
	}
}

func (m *Model) InitFeedingList(chainDBID uint) *Model {
	m.Db.Where("chain_db_id = ? AND status = 2", chainDBID).Find(&m.List)
	return m
}

func (m *Model) ListEmpty() bool {
	return len(m.List) == 0
}

func (m *Model) SetFeedFinish() {
	m.Db.Model(&m.Data).Update("status", 0)
}

func (m *Model) InitWaitingList(chainDBID uint) *Model {
	m.Db.Where("chain_db_id = ? AND status = 1", chainDBID).Find(&m.List)
	return m
}

func (m *Model) SetFeeding(id uint) {
	m.Db.Model(&m.Data).Updates(map[string]any{
		"status":    2,
		"module_id": id,
	})
}

func (m *Model) SetCollectFinish() {
	m.Db.Model(&m.Data).Update("collect_status", 0)
}

func (m *Model) SetCollecting(id uint) {
	m.Db.Model(&m.Data).Updates(map[string]any{
		"collect_status": 1,
		"module_id":      id,
	})
}

func (m *Model) SetWaitFeed() {
	m.Db.Model(&m.Data).Update("status", 1)
}
