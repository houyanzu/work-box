package mnemonics

import (
	"errors"
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/crypto"
	"github.com/houyanzu/work-box/lib/mnemonic"
)

// BoxMnemonics 助记词表结构
type BoxMnemonics struct {
	ID     uint   `gorm:"column:id;primaryKey;autoIncrement"`
	Words  []byte `gorm:"column:words;type:varbinary(255);not null"`
	Remark string `gorm:"column:remark;type:varchar(255);default:''"`
}

// TableName 指定表名
func (BoxMnemonics) TableName() string {
	return "box_mnemonics"
}

var haveTable = false

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_mnemonics` (\n  `id` int unsigned NOT NULL AUTO_INCREMENT,\n  `words` varbinary(255) NOT NULL,\n  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '',\n  PRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci"
	return db.Exec(sql).Error
}

type Model struct {
	*database.MysqlContext
	Data    BoxMnemonics
	List    []BoxMnemonics
	Decoded bool
	Total   int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxMnemonics, 0)
	data := BoxMnemonics{}
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

	return &Model{ctx, data, list, false, 0}
}

func (m *Model) Add() {
	m.Db.Create(&m.Data)
}

func (m *Model) InitByID(ID uint) *Model {
	m.Db.Take(&m.Data, ID)
	return m
}

func (m *Model) GetWords(password []byte, de crypto.Decoder) (priKey string) {
	var temp []byte
	copy(temp, m.Data.Words)
	de.SetBytes(temp)
	en := de.Decode(password)

	return en
}

func (m *Model) DecodeWords(password []byte, de crypto.Decoder) {
	words := m.GetWords(password, de)
	m.Data.Words = []byte(words)
	m.Decoded = true
}

func (m *Model) GetAddressAndPriKeyByIndex(index int) (address string, priKey string, err error) {
	if !m.Decoded {
		return "", "", errors.New("please decode words first")
	}
	mn, err := mnemonic.NewMnemonic(string(m.Data.Words))
	if err != nil {
		return "", "", err
	}
	return mn.GetAddressAndPrivateKeyByIndex(index)
}
