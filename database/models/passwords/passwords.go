package passwords

import (
	"github.com/houyanzu/work-box/database"
	"github.com/houyanzu/work-box/lib/crypto"
	"time"
)

type BoxPasswords struct {
	ID       uint   `json:"id" gorm:"column:id"`
	Module   string `json:"module" gorm:"column:module"`
	Password []byte `json:"password" gorm:"column:password"`
}

func (m *BoxPasswords) TableName() string {
	return "box_passwords"
}

var haveTable = false

func createTable() error {
	db := database.GetDB()
	sql := "CREATE TABLE `box_passwords` (\n\t`id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT,\n\t`module` varchar(32) NOT NULL,\n\t`password` varbinary(255) NOT NULL,\n\tPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8;"
	return db.Exec(sql).Error
}

type Model struct {
	*database.MysqlContext
	Data  BoxPasswords
	List  []BoxPasswords
	Total int64
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	list := make([]BoxPasswords, 0)
	data := BoxPasswords{}
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

func (m *Model) Exists() bool {
	return m.Data.ID > 0
}

func (m *Model) Create(module string, password []byte, en crypto.Encoder) string {
	str := crypto.Sha256Str(time.Now().Format(time.RFC3339) + module)
	pass := str[:16]

	en.SetString(pass)
	m.Data.Password = en.Encode(password)
	m.Data.Module = module
	m.Add()

	return pass
}

func (m *Model) InitByModule(module string) *Model {
	m.Db.Where("module = ?", module).Take(&m.Data)
	return m
}

func (m *Model) GetPassword(password []byte, de crypto.Decoder) []byte {
	de.SetBytes(m.Data.Password)
	pass := de.Decode(password)
	return []byte(pass)
}
