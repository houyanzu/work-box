package asset

import (
	"errors"
	"fmt"
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/database/models/assetrecord"
	"github.com/houyanzu/work-box/database/models/tokengroup"
	"github.com/houyanzu/work-box/lib/mytime"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BoxAsset struct {
	ID            uint
	UserId        uint
	TokenGroupId  uint
	Symbol        string
	Balance       decimal.Decimal
	FreezeBalance decimal.Decimal
	UpdateTime    mytime.DateTime
}

var haveTable = false

func (data *BoxAsset) BeforeCreate(tx *gorm.DB) error {
	data.UpdateTime = mytime.NewFromNow()
	return nil
}

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_asset` (\n\t`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`user_id` int(11) UNSIGNED NOT NULL,\n\t`token_group_id` int(11) UNSIGNED NOT NULL,\n\t`symbol` varchar(10) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`balance` decimal(32,0)  UNSIGNED NOT NULL DEFAULT 0,\n\t`freeze_balance` decimal(32,0)  UNSIGNED NOT NULL DEFAULT 0,\n\t`update_time` datetime NOT NULL,\n\tPRIMARY KEY (`id`),\n\tUnique KEY `ut`(`user_id`,`token_group_id`) USING BTREE\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci\nAUTO_INCREMENT=1\nROW_FORMAT=DYNAMIC\nAVG_ROW_LENGTH=0;"
	return db.Exec(sql).Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxAsset
	List  []BoxAsset
	Total int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxAsset, 0)
	data := BoxAsset{}
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

func NewFromUserIdAndTokenGroupId(ctx *database.MysqlContext, userId, tokenGroupID uint) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	m := New(ctx)
	m.Db.Where("user_id = ? AND token_group_id = ?", userId, tokenGroupID).Take(&m.Data)
	if !m.Exists() {
		tokenGroup := tokengroup.New(nil).InitById(tokenGroupID)
		m.Data.UserId = userId
		m.Data.TokenGroupId = tokenGroupID
		m.Data.Symbol = tokenGroup.Data.Symbol
		m.Data.Balance = decimal.Zero
		m.Data.FreezeBalance = decimal.Zero
		m.Add()
	}
	m.Db.Where("user_id = ? AND token_group_id = ?", userId, tokenGroupID).Take(&m.Data)
	return m
}

func (m *Model) Exists() bool {
	return m.Data.ID > 0
}

func (m *Model) Add() {
	m.Db.Create(&m.Data)
}

func (m *Model) InitListByUserId(userId uint) *Model {
	m.Db.Where("user_id = ?", userId).Find(&m.List)
	return m
}

func (m *Model) GetAvailableBalance() decimal.Decimal {
	return m.Data.Balance.Sub(m.Data.FreezeBalance)
}

func (m *Model) Freeze(value decimal.Decimal) error {
	if !m.Exists() {
		return errors.New("wrong")
	}
	if value.LessThanOrEqual(decimal.Zero) {
		return errors.New("zero")
	}

	sql := fmt.Sprintf("UPDATE `box_asset` SET `freeze_balance` = `freeze_balance` + %s WHERE `id` = %d AND `balance` - `freeze_balance` >= %s;",
		value.String(), m.Data.ID, value.String())

	res := m.Db.Exec(sql).RowsAffected
	if res == 0 {
		return errors.New("insufficient funds")
	}
	return nil
}

func (m *Model) Unfreeze(value decimal.Decimal) error {
	if !m.Exists() {
		return errors.New("wrong")
	}
	if value.LessThanOrEqual(decimal.Zero) {
		return errors.New("zero")
	}

	sql := fmt.Sprintf("UPDATE `box_asset` SET `freeze_balance` = `freeze_balance` - %s WHERE `id` = %d;",
		value.String(), m.Data.ID)
	res := m.Db.Exec(sql).RowsAffected
	if res == 0 {
		return errors.New("exceeds")
	}
	return nil
}

func (m *Model) UnfreezeAndSubBalance(
	module string,
	moduleId uint,
	amount decimal.Decimal,
	remark string,
) error {
	if !m.Exists() {
		return errors.New("wrong")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("zero")
	}

	sql := fmt.Sprintf("UPDATE `box_asset` SET `freeze_balance` = `freeze_balance` - %s, `balance` = `balance` - %s WHERE `id` = %d AND `balance` >= %s AND `freeze_balance` >= %s;",
		amount.String(), amount.String(), m.Data.ID, amount.String(), amount.String())
	res := m.Db.Exec(sql).RowsAffected
	if res == 0 {
		return errors.New("insufficient funds")
	}

	assetRecord := assetrecord.New(m.MysqlContext)
	assetRecord.Data.UserId = m.Data.UserId
	assetRecord.Data.Module = module
	assetRecord.Data.ModuleId = moduleId
	assetRecord.Data.TokenGroupId = m.Data.TokenGroupId
	assetRecord.Data.Symbol = m.Data.Symbol
	assetRecord.Data.Amount = amount
	assetRecord.Data.CreateTime = mytime.NewFromNow()
	assetRecord.Data.Type = 2
	assetRecord.Data.Remark = remark
	assetRecord.Add()
	return nil
}

func (m *Model) SubBalance(
	module string,
	moduleId uint,
	amount decimal.Decimal,
	remark string,
) error {
	if !m.Exists() {
		return errors.New("wrong")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("zero")
	}

	sql := fmt.Sprintf("UPDATE `box_asset` SET `balance` = `balance` - %s WHERE `id` = %d AND `balance` >= %s;",
		amount.String(), amount.String(), m.Data.ID, amount.String())
	res := m.Db.Exec(sql).RowsAffected
	if res == 0 {
		return errors.New("insufficient funds")
	}

	assetRecord := assetrecord.New(m.MysqlContext)
	assetRecord.Data.UserId = m.Data.UserId
	assetRecord.Data.Module = module
	assetRecord.Data.ModuleId = moduleId
	assetRecord.Data.TokenGroupId = m.Data.TokenGroupId
	assetRecord.Data.Symbol = m.Data.Symbol
	assetRecord.Data.Amount = amount
	assetRecord.Data.CreateTime = mytime.NewFromNow()
	assetRecord.Data.Type = 2
	assetRecord.Data.Remark = remark
	assetRecord.Add()
	return nil
}

func (m *Model) AddBalance(
	module string,
	moduleId uint,
	amount decimal.Decimal,
	remark string,
) error {
	if !m.Exists() {
		return errors.New("wrong")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("zero")
	}

	sql := fmt.Sprintf("UPDATE `box_asset` SET `balance`  = `balance` + %s WHERE `id` = %d;",
		amount.String(), m.Data.ID)
	m.Db.Exec(sql)

	assetRecord := assetrecord.New(m.MysqlContext)
	assetRecord.Data.UserId = m.Data.UserId
	assetRecord.Data.Module = module
	assetRecord.Data.ModuleId = moduleId
	assetRecord.Data.TokenGroupId = m.Data.TokenGroupId
	assetRecord.Data.Symbol = m.Data.Symbol
	assetRecord.Data.Amount = amount
	assetRecord.Data.Remark = remark
	assetRecord.Data.CreateTime = mytime.NewFromNow()
	assetRecord.Data.Type = 1
	assetRecord.Add()
	return nil
}
