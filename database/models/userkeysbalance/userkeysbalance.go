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
	ID      uint            `json:"id" gorm:"column:id"`
	KeyID   uint            `json:"key_id" gorm:"column:key_id"`
	TokenID uint            `json:"token_id" gorm:"column:token_id"`
	Balance decimal.Decimal `json:"balance" gorm:"column:balance"`
}

func (m *BoxUserKeysBalance) TableName() string {
	return "box_user_keys_balance"
}

var haveTable = false

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_user_keys_balance` (\n\t`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`key_id` int(11) UNSIGNED NOT NULL,\n\t`token_id` int(11) UNSIGNED NOT NULL,\n\t`balance` decimal(32,0)  NOT NULL,\n\tPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci\nROW_FORMAT=DYNAMIC\nAVG_ROW_LENGTH=0;"
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
	m.Db.Where("key_id = ? AND token_id = ?", keyID, tokenID).Take(&m.Data)
	if !m.Exists() {
		m.Data.KeyID = keyID
		m.Data.TokenID = tokenID
	}
	return m
}

func (m *Model) UpdateBalance() {
	token := tokens.New(nil).InitById(m.Data.TokenID)
	userKey := userkeys.New(nil).InitById(m.Data.KeyID)
	m.Data.Balance, _ = eth.BalanceOf(token.Data.Contract, userKey.Data.Address)
	m.Db.Save(&m.Data)
}

func (m *Model) InitListByBalance(tokenID uint, balance decimal.Decimal) *Model {
	m.Db.Where("token_id = ?", tokenID).Where(fmt.Sprintf("balance >= %s", balance.String())).Find(&m.List)
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
