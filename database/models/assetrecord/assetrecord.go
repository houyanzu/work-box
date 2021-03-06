package assetrecord

import (
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/mytime"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BoxAssetRecord struct {
	ID         uint
	UserId     uint
	TokenId    uint
	Symbol     string
	Module     string
	ModuleId   uint
	Amount     decimal.Decimal
	Remark     string
	Type       int8
	CreateTime mytime.DateTime
}

var haveTable = false

func (data *BoxAssetRecord) BeforeCreate(tx *gorm.DB) error {
	data.CreateTime = mytime.NewFromNow()
	return nil
}

func createTable() error {
	db := database.GetDB()
	return db.Exec("CREATE TABLE `box_asset_record` (\n\t`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`user_id` int(11) UNSIGNED NOT NULL,\n\t`token_id` int(11) UNSIGNED NOT NULL,\n\t`symbol` varchar(10) NOT NULL,\n\t`module` varchar(32) NOT NULL,\n\t`module_id` int(11) UNSIGNED NOT NULL,\n\t`amount` decimal(32,0)  UNSIGNED NOT NULL DEFAULT 0,\n\t`remark` varchar(255) NOT NULL,\n\t`type` tinyint(1) NOT NULL,\n\t`create_time` datetime NOT NULL,\n\tPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8;").Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxAssetRecord
	List  []BoxAssetRecord
	Total int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxAssetRecord, 0)
	data := BoxAssetRecord{}
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
