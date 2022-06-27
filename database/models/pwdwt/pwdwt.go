package pwdwt

import (
	"github.com/houyanzu/work-box/database"
)

type BoxPasswordWrongTimes struct {
	ID    uint
	Times uint
}

type Model struct {
	*database.MysqlContext
	Data BoxPasswordWrongTimes
}

var haveTable = false

func createTable() error {
	db := database.GetDB()
	err := db.Exec("CREATE TABLE `box_password_wrong_times` (\n\t`id` tinyint(1) UNSIGNED NOT NULL,\n\t`times` tinyint(1) UNSIGNED NOT NULL DEFAULT 0,\n\tPRIMARY KEY (`id`)\n) ENGINE=InnoDB\nDEFAULT CHARACTER SET=utf8;").Error
	if err != nil {
		return err
	}
	db.Create(&BoxPasswordWrongTimes{1, 0})
	return nil
}

func New(ctx *database.MysqlContext) *Model {
	if ctx == nil {
		ctx = database.GetContext()
	}
	data := BoxPasswordWrongTimes{}
	if !haveTable {
		hasTable := ctx.Db.Migrator().HasTable(&data)
		if !hasTable {
			err := createTable()
			if err != nil {
				panic(err)
			}
			data.ID = 1
			ctx.Db.Create(&data)
		}
		haveTable = true
	}

	return &Model{ctx, data}
}

func (m *Model) Wrong() {
	m.Db.Exec("UPDATE `password_wrong_times` SET `times` = `times` + 1 WHERE `id` = 1;")
}

func (m *Model) GetTimes() uint {
	m.Db.Take(&m.Data, 1)
	return m.Data.Times
}

func (m *Model) ResetTimes() {
	m.Db.Exec("UPDATE `password_wrong_times` SET `times` = 0 WHERE `id` = 1;")
}
