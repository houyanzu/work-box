package locktransferdetails

import (
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/mytime"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"strings"
)

type BoxLockTransferDetails struct {
	ID               uint
	ChainDbId        uint
	Module           string
	LockContract     string
	Token            string
	To               string
	Amount           decimal.Decimal
	ReleaseStartTime uint
	ReleaseCycle     uint
	ReleaseTimes     uint
	Status           int8
	TransferId       uint
	CreateTime       mytime.DateTime
}

var haveTable = false

func (l *BoxLockTransferDetails) BeforeCreate(tx *gorm.DB) error {
	l.LockContract = strings.ToLower(l.LockContract)
	l.Token = strings.ToLower(l.Token)
	l.To = strings.ToLower(l.To)
	l.CreateTime = mytime.NewFromNow()
	return nil
}

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_lock_transfer_details` (\n\t`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`chain_db_id` int(11) UNSIGNED NOT NULL DEFAULT 1,\n\t`module` varchar(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`lock_contract` char(42) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`token` char(42) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`to` char(42) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`amount` decimal(32,0)  UNSIGNED NOT NULL DEFAULT 0,\n\t`release_start_time` int(11) UNSIGNED NOT NULL,\n\t`release_cycle` int(11) UNSIGNED NOT NULL,\n\t`release_times` int(11) UNSIGNED NOT NULL,\n\t`status` tinyint(1) NOT NULL,\n\t`transfer_id` int(11) NOT NULL,\n\t`create_time` datetime NOT NULL,\n\tPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci\nROW_FORMAT=DYNAMIC\nAVG_ROW_LENGTH=0;"
	return db.Exec(sql).Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxLockTransferDetails
	List  []BoxLockTransferDetails
	Total int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxLockTransferDetails, 0)
	data := BoxLockTransferDetails{}
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

func (m *Model) Exists() bool {
	return m.Data.ID > 0
}

func (m *Model) InitByData(data BoxLockTransferDetails) *Model {
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

func (m *Model) InitWaitingList(chainDBID uint, limit int, module string) *Model {
	m.Error = m.Db.Where("chain_db_id = ? AND module = ? AND status = 0", chainDBID, module).Limit(limit).Find(&m.List).Error
	return m
}

func (m *Model) SetExec(ids []uint, transferId uint) bool {
	if m.Data.ID > 0 {
		return false
	}
	m.Error = m.Db.Model(&m.Data).Where("id in (?)", ids).Updates(map[string]any{
		"status":      1,
		"transfer_id": transferId,
	}).Error
	return true
}

func (m *Model) SetSuccess(transferId uint) bool {
	if m.Data.ID > 0 {
		return false
	}
	m.Error = m.Db.Model(&m.Data).Where("transfer_id = ?", transferId).Updates(map[string]any{
		"status": 2,
	}).Error
	return true
}

func (m *Model) SetFail(transferId uint) bool {
	if m.Data.ID > 0 {
		return false
	}
	m.Error = m.Db.Model(&m.Data).Where("transfer_id = ?", transferId).Updates(map[string]any{
		"status": -1,
	}).Error
	return true
}

func (m *Model) Reset(transferId uint) bool {
	if m.Data.ID > 0 {
		return false
	}
	m.Error = m.Db.Model(&m.Data).Where("transfer_id = ?", transferId).Updates(map[string]any{
		"status": 0,
	}).Error
	return true
}
