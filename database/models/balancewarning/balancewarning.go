package balancewarning

import (
	"github.com/houyanzu/work-box/database"
	"github.com/shopspring/decimal"
)

type BoxBalanceWarning struct {
	ID             uint64          `json:"id" gorm:"column:id"`
	Address        string          `json:"address" gorm:"column:address"`
	ChainDBID      uint            `json:"chain_db_id" gorm:"column:chain_db_id"`
	Token          string          `json:"token" gorm:"column:token"`
	WarningBalance decimal.Decimal `json:"warning_balance" gorm:"column:warning_balance"`
	Status         uint8           `json:"status" gorm:"column:status"`
	Remark         string          `json:"remark" gorm:"column:remark"`
}

func (m *BoxBalanceWarning) TableName() string {
	return "box_balance_warning"
}

var haveTable = false

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_balance_warning` (\n\t`id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`address` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`chain_db_id` int(10) UNSIGNED NOT NULL,\n\t`token` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`warning_balance` decimal(64,0)  UNSIGNED NOT NULL,\n\t`status` tinyint(1) UNSIGNED NOT NULL COMMENT '0-关闭，1-打开',\n\t`remark` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '备注',\n\tPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci\nAUTO_INCREMENT=1\nROW_FORMAT=COMPACT\nAVG_ROW_LENGTH=0;"
	return db.Exec(sql).Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxBalanceWarning
	List  []BoxBalanceWarning
	Total int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxBalanceWarning, 0)
	data := BoxBalanceWarning{}
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

func (m *Model) GetAllOpen() *Model {
	m.Db.Where("status = 1").Find(&m.List)
	return m
}

func (m *Model) Foreach(f func(key int, value *Model) (b bool)) {
	for i, v := range m.List {
		if f(i, &Model{m.MysqlContext, v, nil, 0}) {
			break
		}
	}
}
