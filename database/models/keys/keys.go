package keys

import (
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/crypto"
	"github.com/houyanzu/work-box/lib/mytime"
	"github.com/houyanzu/work-box/tool/eth"
	"gorm.io/gorm"
	"strings"
)

type BoxKeys struct {
	ID         uint            `json:"id" gorm:"column:id"`
	Address    string          `json:"address" gorm:"column:address"`
	PriKey     []byte          `json:"pri_key" gorm:"column:pri_key"`
	Remark     string          `json:"remark" gorm:"column:remark"`
	CreateTime mytime.DateTime `json:"create_time" gorm:"column:create_time"`
}

func (m *BoxKeys) TableName() string {
	return "box_keys"
}

func (data *BoxKeys) BeforeCreate(tx *gorm.DB) error {
	data.CreateTime = mytime.NewFromNow()
	return nil
}

var haveTable = false

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_keys` (\n\t`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`address` char(42) NOT NULL,\n\t`pri_key` varbinary(255) NULL,\n\t`remark` varchar(64) NOT NULL,\n\t`create_time` datetime NOT NULL,\n\tPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8;"
	return db.Exec(sql).Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxKeys
	List  []BoxKeys
	Total int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxKeys, 0)
	data := BoxKeys{}
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

func (m *Model) CreateKey(password, remark string, en crypto.Encoder) (*Model, error) {
	addr, priKey, err := eth.CreateAddress()
	if err != nil {
		return nil, err
	}

	en.SetString(priKey)
	de := en.Encode([]byte(password))
	m.Data.Address = strings.ToLower(addr)
	m.Data.PriKey = de
	m.Data.Remark = remark
	m.Add()
	return m, nil
}

func (m *Model) Add() {
	m.Db.Create(&m.Data)
}

func (m *Model) InitByID(ID uint) *Model {
	m.Db.Take(&m.Data, ID)
	return m
}

func (m *Model) GetPriKey(password []byte, de crypto.Decoder) (priKey string) {
	de.SetBytes(m.Data.PriKey)
	en := de.Decode(password)
	return en
}
