package transferrecords

import (
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/mytime"
	"gorm.io/gorm"
	"strings"
)

type BoxTransferRecords struct {
	ID         uint
	Module     string
	Type       int8
	From       string
	Hash       string
	Status     int8
	Nonce      uint64
	CreateTime mytime.DateTime
}

var haveTable = false

func (c *BoxTransferRecords) BeforeCreate(tx *gorm.DB) error {
	c.From = strings.ToLower(c.From)
	c.Status = 1
	c.CreateTime = mytime.NewFromNow()
	return nil
}

func createTable() error {
	db := database.GetDB()
	return db.Exec("CREATE TABLE `box_transfer_records` (\n\t`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`module` varchar(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`type` tinyint(1) NOT NULL,\n\t`from` char(42) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '',\n\t`hash` char(66) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`nonce` int(11) UNSIGNED NOT NULL,\n\t`status` tinyint(1) NOT NULL,\n\t`create_time` datetime NOT NULL,\n\tPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci\nAUTO_INCREMENT=1\nROW_FORMAT=DYNAMIC\nAVG_ROW_LENGTH=0;").Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxTransferRecords
	List  []BoxTransferRecords
	Total int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxTransferRecords, 0)
	data := BoxTransferRecords{}
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

func (m *Model) InitByUserData(data BoxTransferRecords) *Model {
	m.Data = data
	return m
}

func (m *Model) Foreach(f func(index int, m *Model)) {
	for k, v := range m.List {
		mm := New(nil).InitByUserData(v)
		f(k, mm)
	}
}

func (m *Model) Add() {
	m.Error = m.Db.Create(&m.Data).Error
}

func (m *Model) InitPending(from string, module string) *Model {
	m.Error = m.Db.Where("`from` = ? AND `module` = ? AND `status` = 1", strings.ToLower(from), module).
		Take(&m.Data).Error
	return m
}

func (m *Model) InitByID(ID uint) *Model {
	m.Db.Take(&m.Data, ID)
	return m
}

func (m *Model) SetSuccess() bool {
	m.Error = m.Db.Model(&m.Data).Updates(map[string]any{
		"status": 2,
	}).Error
	return true
}

func (m *Model) SetFail() bool {
	m.Error = m.Db.Model(&m.Data).Updates(map[string]any{
		"status": -1,
	}).Error
	return true
}

func (m *Model) Exists() bool {
	return m.Data.ID > 0
}
