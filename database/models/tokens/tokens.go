package tokens

import (
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/mytime"
	"gorm.io/gorm"
	"strings"
)

type BoxTokens struct {
	ID         uint
	Contract   string
	Symbol     string
	Decimals   uint8
	Remark     string
	CreateTime mytime.DateTime
}

var haveTable = false

func (data *BoxTokens) BeforeCreate(tx *gorm.DB) error {
	data.Contract = strings.ToLower(data.Contract)
	data.CreateTime = mytime.NewFromNow()
	return nil
}

func createTable() error {
	db := database.GetDB()
	return db.Exec("CREATE TABLE `box_tokens` (\n\t`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`contract` char(42) NOT NULL,\n\t`symbol` varchar(10) NOT NULL,\n\t`decimals` tinyint(4) NOT NULL,\n\t`remark` text NOT NULL,\n\t`create_time` datetime NOT NULL,\n\tPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8;").Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxTokens
	List  []BoxTokens
	Total int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxTokens, 0)
	data := BoxTokens{}
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

func (m *Model) InitById(id uint) *Model {
	m.Db.Take(&m.Data, id)
	return m
}

func (m *Model) Add() {
	m.Db.Create(&m.Data)
}

func (m *Model) Save() {
	m.Db.Save(&m.Data)
}
