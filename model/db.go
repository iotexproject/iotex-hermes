package model

import (
	"crypto/rsa"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	"github.com/iotexproject/iotex-hermes/key"
	"github.com/iotexproject/iotex-hermes/util"
)

type DBLog struct{}

func (dblog *DBLog) Print(v ...interface{}) {
	fmt.Println(v...)
}

var (
	db         *gorm.DB
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

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

func loadKeys() (err error) {
	if privateKey == nil {
		privateKey, err = key.LoadPrivateKey(util.MustFetchNonEmptyParam("RSA_PRIVATE"))
		if err != nil {
			return fmt.Errorf("load private key error: %v\n", err)
		}
	}
	if publicKey == nil {
		publicKey, err = key.LoadPublicKey(util.MustFetchNonEmptyParam("RSA_PUBLIC"))
		if err != nil {
			return fmt.Errorf("load public key error: %v", err)
		}
	}

	return nil
}

func DB() *gorm.DB {
	return db.New()
}

func Begin() *gorm.DB {
	return db.Begin()
}
