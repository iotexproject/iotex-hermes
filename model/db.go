package model

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type DBLog struct{}

func (dblog *DBLog) Print(v ...interface{}) {
	fmt.Println(v...)
}

var db *gorm.DB

func init() {
	if db != nil {
		return
	}

	dataSourceName := os.Getenv("DB_MYSQL")

	var err error
	db, err = gorm.Open("mysql", dataSourceName)
	if err != nil {
		fmt.Println(err)
		return
	}

	if os.Getenv("LOG_MODE") == "debug" {
		db.LogMode(true)
		db.SetLogger(&DBLog{})
	}

	return
}

func DB() *gorm.DB {
	return db.New()
}

func Begin() *gorm.DB {
	return db.Begin()
}
