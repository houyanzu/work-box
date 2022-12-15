package chainrecord

import (
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/mytime"
	"gorm.io/gorm"
	"strings"
)

type BoxMonitorChainRecord struct {
	ID         uint
	ChainDbId  uint
	Contract   string
	BlockNum   uint64
	EventId    string
	Hash       string
	CreateTime mytime.DateTime
}

var haveTable = false

func (c *BoxMonitorChainRecord) BeforeCreate(tx *gorm.DB) error {
	c.Contract = strings.ToLower(c.Contract)
	c.CreateTime = mytime.NewFromNow()
	return nil
}

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_monitor_chain_record` (\n\t`id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`chain_db_id` int(11) UNSIGNED NOT NULL DEFAULT 1,\n\t`contract` char(42) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`block_num` bigint(20) UNSIGNED NOT NULL,\n\t`event_id` char(66) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`hash` char(66) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`create_time` datetime NOT NULL,\n\tPRIMARY KEY (`id`),\n\tKEY `cb`(`contract`,`block_num`) USING BTREE\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci\nAUTO_INCREMENT=1\nROW_FORMAT=DYNAMIC\nAVG_ROW_LENGTH=0;"
	return db.Exec(sql).Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxMonitorChainRecord
	List  []BoxMonitorChainRecord
	Total int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxMonitorChainRecord, 0)
	data := BoxMonitorChainRecord{}
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

func (m *Model) InitByID(ID uint) *Model {
	m.Db.Take(&m.Data, ID)
	return m
}

func (m *Model) InitByData(data BoxMonitorChainRecord) *Model {
	m.Data = data
	return m
}

func (m *Model) Foreach(f func(index int, m *Model)) {
	for k, v := range m.List {
		mm := New(nil).InitByData(v)
		f(k, mm)
	}
}

func (m *Model) Add() {
	m.Error = m.Db.Create(&m.Data).Error
}

func GetLastBlockNum(chainDBID uint, contract string) uint64 {
	db := database.GetDB()
	var c BoxMonitorChainRecord
	db.Where("chain_db_id = ? AND contract = ?", chainDBID, strings.ToLower(contract)).Order("block_num desc").Take(&c)
	return c.BlockNum
}
