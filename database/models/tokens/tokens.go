package tokens

import (
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/mytime"
	"gorm.io/gorm"
	"strings"
)

type BoxTokens struct {
	ID         uint            `json:"id" gorm:"column:id"`
	ChainDbID  uint            `json:"chain_db_id" gorm:"column:chain_db_id"`
	GroupID    uint            `json:"group_id" gorm:"column:group_id"`
	Contract   string          `json:"contract" gorm:"column:contract"`
	Symbol     string          `json:"symbol" gorm:"column:symbol"`
	Decimals   int8            `json:"decimals" gorm:"column:decimals"`
	Remark     string          `json:"remark" gorm:"column:remark"`
	CreateTime mytime.DateTime `json:"create_time" gorm:"column:create_time"`
}

func (m *BoxTokens) TableName() string {
	return "box_tokens"
}

var haveTable = false

func (data *BoxTokens) BeforeCreate(tx *gorm.DB) error {
	data.Contract = strings.ToLower(data.Contract)
	data.CreateTime = mytime.NewFromNow()
	return nil
}

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_tokens` (\n\t`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`chain_db_id` int(11) UNSIGNED NOT NULL DEFAULT 1,\n\t`group_id` int(11) UNSIGNED NOT NULL DEFAULT 0,\n\t`contract` char(42) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`symbol` varchar(10) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`decimals` tinyint(4) NOT NULL,\n\t`remark` text CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`create_time` datetime NOT NULL,\n\tPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci\nAUTO_INCREMENT=1\nROW_FORMAT=DYNAMIC\nAVG_ROW_LENGTH=0;"
	return db.Exec(sql).Error
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

func (m *Model) InitByContract(contract string) *Model {
	m.Db.Where("contract = ?", strings.ToLower(contract)).Take(&m.Data)
	return m
}

func (m *Model) Add() {
	m.Db.Create(&m.Data)
}

func (m *Model) Save() {
	m.Db.Save(&m.Data)
}

func (m *Model) InitByData(data BoxTokens) *Model {
	m.Data = data
	return m
}

func (m *Model) Foreach(f func(index int, m *Model)) {
	for k, v := range m.List {
		mm := m.InitByData(v)
		f(k, mm)
	}
}

func (m *Model) Exists() bool {
	return m.Data.ID > 0
}
