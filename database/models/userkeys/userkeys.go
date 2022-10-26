package userkeys

import (
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/database/models/passwords"
	"github.com/houyanzu/work-box/lib/crypto"
	"github.com/houyanzu/work-box/lib/mytime"
	"github.com/houyanzu/work-box/tool/eth"
	"gorm.io/gorm"
	"strings"
	"sync"
)

type BoxUserKeys struct {
	ID               uint            `json:"id" gorm:"column:id"`
	UserID           uint            `json:"user_id" gorm:"column:user_id"`
	Address          string          `json:"address" gorm:"column:address"`
	PrivateKey       []byte          `json:"private_key" gorm:"column:private_key"`
	EthBalance       float64         `json:"eth_balance" gorm:"column:eth_balance"`
	Status           int8            `json:"status" gorm:"status"`
	CollectStatus    int8            `json:"collect_status" gorm:"collect_status"`
	TransferDetailID uint            `json:"transfer_detail_id" gorm:"transfer_detail_id"`
	CreateTime       mytime.DateTime `json:"create_time" gorm:"column:create_time"`
}

func (m *BoxUserKeys) TableName() string {
	return "box_user_keys"
}

func (data *BoxUserKeys) BeforeCreate(tx *gorm.DB) error {
	data.CreateTime = mytime.NewFromNow()
	return nil
}

var haveTable = false

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_user_keys` (\n\t`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`user_id` int(11) UNSIGNED NOT NULL,\n\t`address` char(42) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,\n\t`private_key` varbinary(255) NOT NULL,\n\t`eth_balance` decimal(32,0)  NOT NULL,\n\t`status` tinyint(1) NOT NULL COMMENT '0-正常，1-待转eth，2-转eth中',\n\t`collect_status` tinyint(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '0-正常，1-归集中',\n\t`transfer_detail_id` int(11) UNSIGNED NOT NULL,\n\t`create_time` datetime NOT NULL,\n\tPRIMARY KEY (`id`),\n\tKEY `address`(`address`) USING BTREE\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8 COLLATE=utf8_general_ci\nAUTO_INCREMENT=1\nROW_FORMAT=DYNAMIC\nAVG_ROW_LENGTH=0;"
	return db.Exec(sql).Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxUserKeys
	List  []BoxUserKeys
	Total int64
	mu    sync.Mutex
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxUserKeys, 0)
	data := BoxUserKeys{}
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

	return &Model{ctx, data, list, 0, sync.Mutex{}}
}

func (m *Model) Add() {
	m.Db.Create(&m.Data)
}

func (m *Model) Exists() bool {
	return m.Data.ID > 0
}

func (m *Model) CreateKeys(count int, password []byte, en crypto.Encoder, de crypto.Decoder) error {
	passModel := passwords.New(nil).InitByModule("USER_KEY")
	if !passModel.Exists() {
		passModel.Create("USER_KEY", password, en)
	}
	pass := passModel.GetPassword(password, de)

	for i := 0; i < count; i++ {
		addr, priKey, err := eth.CreateAddress()
		if err != nil {
			return err
		}
		en.SetString(priKey)
		dec := en.Encode(pass)
		m.Data.ID = 0
		m.Data.Address = strings.ToLower(addr)
		m.Data.PrivateKey = dec
		m.Add()
	}

	return nil
}

func (m *Model) OfferKey(userID uint) *Model {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Db.Where("user_id = ?", userID).Take(&m.Data)
	if m.Exists() {
		return m
	}
	m.Db.Where("user_id = 0").Order("id asc").Take(&m.Data)
	rows := m.Db.Where("id = ? AND user_id = 0", m.Data.ID).Update("user_id", userID).RowsAffected
	if rows == 1 {
		m.Data.UserID = userID
		return m
	}
	return New(nil)
}

func (m *Model) GetPriKey(password []byte, de crypto.Decoder) (priKey string) {
	passModel := passwords.New(nil).InitByModule("USER_KEY")
	if !passModel.Exists() {
		return ""
	}
	pass := passModel.GetPassword(password, de)

	de.SetBytes(m.Data.PrivateKey)
	en := de.Decode(pass)
	return en
}

func (m *Model) InitByAddress(address string) *Model {
	m.Db.Where("address = ?", strings.ToLower(address)).Take(&m.Data)
	return m
}

func (m *Model) InitByUserID(userID uint) *Model {
	m.Db.Where("user_id = ?", userID).Take(&m.Data)
	return m
}

func (m *Model) InitById(id uint) *Model {
	m.Db.Take(&m.Data, id)
	return m
}

func (m *Model) InitByData(data BoxUserKeys) *Model {
	m.Data = data
	return m
}

func (m *Model) InitWaitingList() *Model {
	m.Db.Where("status = 1").Find(&m.List)
	return m
}

func (m *Model) InitFeedingList() *Model {
	m.Db.Where("status = 2").Find(&m.List)
	return m
}

func (m *Model) ListEmpty() bool {
	return len(m.List) == 0
}

func (m *Model) Foreach(f func(index int, m *Model)) {
	for k, v := range m.List {
		mm := New(nil).InitByData(v)
		f(k, mm)
	}
}

func (m *Model) SetWaitFeed() {
	m.Db.Model(&m.Data).Update("status", 1)
}

func (m *Model) SetFeeding(id uint) {
	m.Db.Model(&m.Data).Updates(map[string]any{
		"status":             2,
		"transfer_detail_id": id,
	})
}

func (m *Model) SetFeedFinish() {
	m.Db.Model(&m.Data).Update("status", 0)
}

func (m *Model) SetCollecting(id uint) {
	m.Db.Model(&m.Data).Updates(map[string]any{
		"collect_status":     1,
		"transfer_detail_id": id,
	})
}

func (m *Model) SetCollectFinish() {
	m.Db.Model(&m.Data).Update("collect_status", 0)
}
