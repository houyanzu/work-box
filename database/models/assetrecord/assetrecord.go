package assetrecord

import (
	"fmt"
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/mytime"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BoxAssetRecord struct {
	ID           uint            `json:"id"`
	UserId       uint            `json:"user_id"`
	TokenGroupId uint            `json:"token_group_id"`
	Symbol       string          `json:"symbol"`
	Module       string          `json:"module"`
	ModuleId     uint            `json:"module_id"`
	Amount       decimal.Decimal `json:"amount"`
	Remark       string          `json:"remark"`
	Type         int8            `json:"type"`
	CreateTime   mytime.DateTime `json:"create_time"`
}

var haveTable = false

func (data *BoxAssetRecord) BeforeCreate(tx *gorm.DB) error {
	data.CreateTime = mytime.NewFromNow()
	return nil
}

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_asset_record` (\n\t`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`user_id` int(11) UNSIGNED NOT NULL,\n\t`token_group_id` int(11) UNSIGNED NOT NULL,\n\t`symbol` varchar(10) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`module` varchar(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`module_id` int(11) UNSIGNED NOT NULL,\n\t`amount` decimal(32,0)  UNSIGNED NOT NULL DEFAULT 0,\n\t`remark` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`type` tinyint(1) NOT NULL,\n\t`create_time` datetime NOT NULL,\n\tPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci\nAUTO_INCREMENT=1\nROW_FORMAT=DYNAMIC\nAVG_ROW_LENGTH=0;"
	return db.Exec(sql).Error
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

func (m *Model) InitByData(data BoxAssetRecord) *Model {
	m.Data = data
	return m
}

func (m *Model) InitByID(ID uint) *Model {
	m.Db.Take(&m.Data, ID)
	return m
}

func (m *Model) InitListByUserIDAndTokenGroupID(userID, tokenGroupID uint, modules string, assetType int8, start, limit int) *Model {
	where := fmt.Sprintf("user_id = %d", userID)
	if tokenGroupID > 0 {
		where = where + fmt.Sprintf(" AND token_group_id = %d", tokenGroupID)
	}
	if modules != "" {
		where = where + fmt.Sprintf(" AND module IN (%s)", modules)
	}
	if assetType > 0 {
		where = where + fmt.Sprintf(" AND type = %d", assetType)
	}
	m.Db.Where(where).Offset(start).Limit(limit).Order("id DESC").Find(&m.List)
	m.Db.Model(&BoxAssetRecord{}).Where(where).Count(&m.Total)
	return m
}
