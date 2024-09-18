package dispatchers

import (
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/crypto"
	"github.com/houyanzu/work-box/tool/eth"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"strings"
	"sync"
)

var mu sync.Mutex
var haveTable = false

type BoxDispatchers struct {
	ID     uint   `json:"id" gorm:"column:id"`
	Addr   string `json:"addr" gorm:"column:addr"`
	PriKey []byte `json:"pri_key" gorm:"column:pri_key"`
	Status int8   `json:"status" gorm:"column:status"` // 0-空闲，1-使用中，-1-待转手续费，-2-停用，-3-有问题，3-转手续费中
	FeedID uint   `json:"feed_id" gorm:"column:feed_id"`
}

func (m *BoxDispatchers) TableName() string {
	return "box_dispatchers"
}

func (data *BoxDispatchers) BeforeCreate(tx *gorm.DB) error {
	data.Addr = strings.ToLower(data.Addr)
	return nil
}

type Model struct {
	*database.MysqlContext
	Data  BoxDispatchers
	List  []BoxDispatchers
	Total int64
}

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_dispatchers` (\n`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n`addr` char(42) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n`pri_key` varbinary(255) NOT NULL,\n`status` tinyint(2) NOT NULL COMMENT '0-空闲，1-使用中，-1-待转手续费，-2-停用，-3-有问题，3-转手续费中',\n`feed_id` int(11) UNSIGNED NOT NULL DEFAULT 0,\nPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8mb3 COLLATE=utf8_general_ci\nAUTO_INCREMENT=1\nROW_FORMAT=DYNAMIC\nAVG_ROW_LENGTH=0;"
	return db.Exec(sql).Error
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxDispatchers, 0)
	data := BoxDispatchers{}
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

func (m *Model) GetNeedFeedList() *Model {
	m.Db.Where("status = -1").Find(&m.List)
	return m
}

func (m *Model) GetNeedCheckList() *Model {
	m.Db.Where("status = -3").Find(&m.List)
	return m
}

func (m *Model) Add() {
	m.Data.ID = 0
	m.Error = m.Db.Create(&m.Data).Error
}

func (m *Model) CreateKeys(count int, password string, en crypto.Encoder) {
	for i := 0; i < count; i++ {
		addr, priKey, err := eth.CreateAddress()
		if err != nil {
			continue
		}

		en.SetString(priKey)
		de := en.Encode([]byte(password))
		m = New(nil)
		m.Data.Addr = strings.ToLower(addr)
		m.Data.PriKey = de
		m.Add()
	}
}

func (m *Model) GetDispatcher(password string, de crypto.Decoder) *Model {
	mu.Lock()
	defer mu.Unlock()
	m.Db.Where("status = 0").Order("id asc").Take(&m.Data)
	if m.Data.ID == 0 {
		return m
	}
	de.SetBytes(m.Data.PriKey)
	m.Data.PriKey = []byte(de.Decode([]byte(password)))
	m.SetBusy()
	return m
}

func (m *Model) SetFree(minBalance decimal.Decimal) {
	balance, err := eth.BalanceAt(1, m.Data.Addr)
	if err != nil {
		m.Db.Model(&m.Data).Update("status", -3)
		return
	}
	if balance.LessThan(minBalance) {
		m.Db.Model(&m.Data).Update("status", -1)
		return
	}
	m.Db.Model(&m.Data).Update("status", 0)
}

func (m *Model) SetNeedFeed() {
	m.Db.Model(&m.Data).Update("status", -1)
}

func (m *Model) SetBusy() {
	m.Db.Model(&m.Data).Update("status", 1)
}

func (m *Model) SetFeeding(ids []uint, feedId uint) {
	updateMap := map[string]interface{}{
		"status":  3,
		"feed_id": feedId,
	}
	m.Db.Table(m.Data.TableName()).Where("id in (?)", ids).Updates(updateMap)
}

func (m *Model) SetFreeByFeedId(feedId uint) {
	m.Db.Table(m.Data.TableName()).Where("feed_id = ?", feedId).Update("status", 0)
}

func (m *Model) SetNeedFeedByFeedId(feedId uint) {
	m.Db.Table(m.Data.TableName()).Where("feed_id = ?", feedId).Update("status", -1)
}

func (m *Model) Exists() bool {
	return m.Data.ID > 0
}

func (m *Model) Foreach(f func(key int, value *Model) bool) {
	for i, v := range m.List {
		if f(i, &Model{m.MysqlContext, v, nil, 0}) {
			break
		}
	}

}
