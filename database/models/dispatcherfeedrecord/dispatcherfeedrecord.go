package dispatcherfeedrecord

import (
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/mytime"
	"gorm.io/gorm"
)

var haveTable = false

type BoxDispatcherFeedRecord struct {
	ID         uint            `json:"id" gorm:"column:id"`
	TrxID      string          `json:"trx_id" gorm:"column:trx_id"`
	Status     int8            `json:"status" gorm:"column:status"`
	Nonce      uint            `json:"nonce" gorm:"column:nonce"`
	CreateTime mytime.DateTime `json:"create_time" gorm:"column:create_time"`
}

func (m *BoxDispatcherFeedRecord) TableName() string {
	return "box_dispatcher_feed_record"
}

func (data *BoxDispatcherFeedRecord) BeforeCreate(tx *gorm.DB) error {
	data.CreateTime = mytime.NewFromNow()
	return nil
}

type Model struct {
	*database.MysqlContext
	Data  BoxDispatcherFeedRecord
	List  []BoxDispatcherFeedRecord
	Total int64
}

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_dispatcher_feed_record` (\n`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n`trx_id` char(66) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n`status` tinyint(4) NOT NULL,\n`nonce` int(11) UNSIGNED NOT NULL,\n`create_time` datetime NOT NULL,\nPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8mb3 COLLATE=utf8_general_ci\nAUTO_INCREMENT=1\nROW_FORMAT=DYNAMIC\nAVG_ROW_LENGTH=0;"
	return db.Exec(sql).Error
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxDispatcherFeedRecord, 0)
	data := BoxDispatcherFeedRecord{}
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
	m.Data.ID = 0
	m.Error = m.Db.Create(&m.Data).Error
}

func (m *Model) SetSuccess() {
	m.Db.Model(&m.Data).Update("status", 2)
}

func (m *Model) SetFail() {
	m.Db.Model(&m.Data).Update("status", -1)
}

func (m *Model) InitPendingList() *Model {
	m.Db.Where("status = 1").Find(&m.List)
	return m
}
