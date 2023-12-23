package main

import (
	"log"

	"github.com/iotexproject/iotex-hermes/cmd/dao"
)

func main() {
	err := dao.ConnectDatabase()
	if err != nil {
		log.Fatalf("create database error: %v\n", err)
	}

	rows, err := dao.FindByStatus("pending")
	if err != nil {
		log.Fatalf("query records by status error: %v\n", err)
	}
	for _, v := range rows {
		v.Signature = ""
		v.Status = "new"
		if err = v.Save(dao.DB()); err != nil {
			log.Fatalf("save record error: %v\n", err)
		}
	}
}
