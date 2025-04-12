package mnemonicaddresses

import (
	"github.com/houyanzu/work-box/database"
	"strings"
)

// BoxMnemonicAddresses 助记词地址表结构
type BoxMnemonicAddresses struct {
	ID         uint   `gorm:"column:id;primaryKey;autoIncrement"`
	Address    string `gorm:"column:address;type:char(42);not null;default:''"`
	MnemonicID uint   `gorm:"column:mnemonic_id;not null"`
	Index      int    `gorm:"column:index;not null"`
	Remark     string `gorm:"column:remark;type:varchar(255);not null;default:''"`
}

// TableName 指定表名
func (b *BoxMnemonicAddresses) TableName() string {
	return "box_mnemonic_addresses"
}

var haveTable = false

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_mnemonic_addresses` (\n  `id` int unsigned NOT NULL AUTO_INCREMENT,\n  `address` char(42) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '',\n  `mnemonic_id` int unsigned NOT NULL,\n  `index` int unsigned NOT NULL,\n  `remark` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '',\n  PRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8"
	return db.Exec(sql).Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxMnemonicAddresses
	List  []BoxMnemonicAddresses
	Total int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxMnemonicAddresses, 0)
	data := BoxMnemonicAddresses{}
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
	m.Data.Address = strings.ToLower(m.Data.Address)
	m.Db.Create(&m.Data)
}

func (m *Model) InitByID(ID uint) *Model {
	m.Db.Take(&m.Data, ID)
	return m
}
