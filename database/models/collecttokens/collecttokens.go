package collecttokens

import (
	"github.com/houyanzu/work-box/database"
	"github.com/shopspring/decimal"
)

type BoxCollectTokens struct {
	ID            uint            `json:"id" gorm:"column:id"`
	TokenID       uint            `json:"token_id" gorm:"column:token_id"`
	Status        int8            `json:"status" gorm:"column:status"`
	FeedAmount    decimal.Decimal `json:"feed_amount" gorm:"column:feed_amount"`
	CollectAmount decimal.Decimal `json:"collect_amount" gorm:"column:collect_amount"`
}

func (m *BoxCollectTokens) TableName() string {
	return "box_collect_tokens"
}

var haveTable = false

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_collect_tokens` (\n\t`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`token_id` int(11) UNSIGNED NOT NULL,\n\t`status` tinyint(1) NOT NULL DEFAULT 1,\n\t`feed_amount` decimal(32,0)  UNSIGNED NOT NULL,\n\t`collect_amount` decimal(32,0)  UNSIGNED NOT NULL,\n\tPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci\nROW_FORMAT=DYNAMIC\nAVG_ROW_LENGTH=0;"
	return db.Exec(sql).Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxCollectTokens
	List  []BoxCollectTokens
	Total int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxCollectTokens, 0)
	data := BoxCollectTokens{}
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

func (m *Model) InitByData(data BoxCollectTokens) *Model {
	m.Data = data
	return m
}

func (m *Model) InitByID(ID uint) *Model {
	m.Db.Take(&m.Data, ID)
	return m
}

func (m *Model) InitByTokenID(tokenID uint) *Model {
	m.Db.Where("token_id = ?", tokenID).Take(&m.Data)
	return m
}

func (m *Model) InitAvaList() *Model {
	m.Db.Where("status = 1").Find(&m.List)
	return m
}

func (m *Model) ListEmpty() bool {
	return len(m.List) == 0
}

func (m *Model) Foreach(f func(index int, m *Model)) {
	for k, v := range m.List {
		mm := New(nil).InitByData(v)
		f(k, mm)
	}
}
