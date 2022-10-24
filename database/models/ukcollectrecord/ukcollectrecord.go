package ukcollectrecord

import (
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/mytime"
	"gorm.io/gorm"
)

type BoxUkCollectRecord struct {
	ID         uint            `json:"id" gorm:"column:id"`
	KeyID      uint            `json:"key_id" gorm:"column:key_id"`
	Hash       string          `json:"hash" gorm:"column:hash"`
	Status     int8            `json:"status" gorm:"column:status"`
	Nonce      uint64          `json:"nonce" gorm:"column:nonce"`
	CreateTime mytime.DateTime `json:"create_time" gorm:"column:create_time"`
}

func (m *BoxUkCollectRecord) TableName() string {
	return "box_uk_collect_record"
}

func (data *BoxUkCollectRecord) BeforeCreate(tx *gorm.DB) error {
	data.CreateTime = mytime.NewFromNow()
	return nil
}

var haveTable = false

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_uk_collect_record` (\n\t`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`key_id` int(11) UNSIGNED NOT NULL,\n\t`hash` char(66) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`status` tinyint(1) NOT NULL,\n\t`nonce` int(11) UNSIGNED NOT NULL,\n\t`create_time` datetime NOT NULL,\n\tPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci\nROW_FORMAT=DYNAMIC\nAVG_ROW_LENGTH=0;"
	return db.Exec(sql).Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxUkCollectRecord
	List  []BoxUkCollectRecord
	Total int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxUkCollectRecord, 0)
	data := BoxUkCollectRecord{}
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

func (m *Model) InitPendingByKeyID(keyID uint) *Model {
	m.Db.Where("key_id = ? AND status = 1", keyID).Take(&m.Data)
	return m
}

func (m *Model) SetSuccess() {
	m.Db.Model(&m.Data).Update("status", 2)
}

func (m *Model) SetFail() {
	m.Db.Model(&m.Data).Update("status", -1)
}
