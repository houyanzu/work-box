package tokengroup

import (
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/mytime"
	"gorm.io/gorm"
)

type BoxTokenGroup struct {
	ID         uint            `json:"id" gorm:"column:id"`
	Symbol     string          `json:"symbol" gorm:"column:symbol"`
	Remark     string          `json:"remark" gorm:"column:remark"`
	CreateTime mytime.DateTime `json:"create_time" gorm:"column:create_time"`
}

func (m *BoxTokenGroup) TableName() string {
	return "box_asset_group"
}

var haveTable = false

func (data *BoxTokenGroup) BeforeCreate(tx *gorm.DB) error {
	data.CreateTime = mytime.NewFromNow()
	return nil
}

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_token_group` (\n`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n`symbol` varchar(32) NOT NULL,\n`remark` varchar(255) NOT NULL,\n`create_time` datetime NOT NULL,\nPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8;"
	return db.Exec(sql).Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxTokenGroup
	List  []BoxTokenGroup
	Total int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxTokenGroup, 0)
	data := BoxTokenGroup{}
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

func (m *Model) InitById(id uint) *Model {
	m.Db.Take(&m.Data, id)
	return m
}
