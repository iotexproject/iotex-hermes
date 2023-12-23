package dao

import (
	"crypto/rsa"
	"fmt"

	"github.com/jinzhu/gorm"
	// mysql dialects
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/iotexproject/iotex-hermes/cmd/key"
	"github.com/iotexproject/iotex-hermes/util"
)

var db *gorm.DB
var privateKey *rsa.PrivateKey
var publicKey *rsa.PublicKey

// ConnectDatabase connect database
func ConnectDatabase() error {
	var err error
	db, err = gorm.Open("mysql", util.MustFetchNonEmptyParam("DB_CONN"))
	if err != nil {
		return fmt.Errorf("open database error: %v", err)
	}
	db.AutoMigrate(&DropRecord{}, &SmallRecord{}, &SmallRecordBak{}, &Account{})

	privateKey, err = key.LoadPrivateKey(util.MustFetchNonEmptyParam("RSA_PRIVATE"))
	if err != nil {
		return fmt.Errorf("load private key error: %v", err)
	}
	publicKey, err = key.LoadPublicKey(util.MustFetchNonEmptyParam("RSA_PUBLIC"))
	if err != nil {
		return fmt.Errorf("load public key error: %v", err)
	}

	return nil
}

// Transaction begin transaction
func Transaction() *gorm.DB {
	return db.Begin()
}

// DB export db
func DB() *gorm.DB {
	return db
}
