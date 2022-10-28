package chains

import (
	"github.com/houyanzu/work-box/database"
)

type BoxChains struct {
	ID      uint   `json:"id" gorm:"column:id"`
	Name    string `json:"name" gorm:"column:name"`
	Rpc     string `json:"rpc" gorm:"column:rpc"`
	ChainID uint   `json:"chain_id" gorm:"column:chain_id"`
	ApiHost string `json:"api_host" gorm:"column:api_host"`
	ApiKey  string `json:"api_key" gorm:"column:api_key"`
}

func (m *BoxChains) TableName() string {
	return "box_chains"
}

var haveTable = false

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_chains` (\n\t`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`name` varchar(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`rpc` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`chain_id` int(11) UNSIGNED NOT NULL,\n\t`api_host` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '',\n\t`api_key` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '',\n\tPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci\nROW_FORMAT=DYNAMIC\nAVG_ROW_LENGTH=0;"
	return db.Exec(sql).Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxChains
	List  []BoxChains
	Total int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxChains, 0)
	data := BoxChains{}
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
