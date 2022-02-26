package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/iotexproject/iotex-hermes/cmd/dao"
)

func main() {
	err := dao.ConnectDatabase()
	if err != nil {
		log.Fatalf("create database error: %v\n", err)
	}

	voters, err := dao.FindVotersByStatus("new")
	if err != nil {
		log.Fatalf("query new voters error: %v\n", err)
	}
	for _, voter := range voters {
		rows, err := dao.FindByVoterAndStatus(voter, "new")
		if err != nil {
			log.Fatalf("query new rewards by voter error: %v\n", err)
		}
		if len(rows) < 2 {
			continue
		}
		tx := dao.Transaction()
		amount, _ := new(big.Int).SetString(rows[0].Amount, 10)
		for i := 1; i < len(rows); i++ {
			temp, _ := new(big.Int).SetString(rows[i].Amount, 10)
			amount = new(big.Int).Add(amount, temp)
			rows[i].Status = fmt.Sprintf("merged-%d", rows[i].ID)
			if err = rows[i].Save(tx); err != nil {
				tx.Rollback()
				log.Fatalf("save merged record error: %v\n", err)
			}
		}
		rows[0].Amount = amount.String()
		if err = rows[0].Save(tx); err != nil {
			tx.Rollback()
			log.Fatalf("save merged to record error: %v\n", err)
		}
		tx.Commit()
	}
}
