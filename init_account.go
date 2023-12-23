package main

import (
	"log"

	"github.com/iotexproject/iotex-antenna-go/v2/account"
	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-hermes/cmd/dao"
	"github.com/iotexproject/iotex-hermes/cmd/distribute"
	"github.com/iotexproject/iotex-hermes/util"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
)

func main() {
	err := dao.ConnectDatabase()
	if err != nil {
		log.Fatalf("create database error: %v\n", err)
	}
	endpoint := util.MustFetchNonEmptyParam("IO_ENDPOINT")
	conn, err := iotex.NewDefaultGRPCConn(endpoint)
	if err != nil {
		log.Fatalf("construct grpc connection error: %v\n", err)
	}
	defer conn.Close()
	emptyAccount, err := account.NewAccount()
	if err != nil {
		log.Fatalf("new empty account error: %v\n", err)
	}
	c := iotex.NewAuthedClient(iotexapi.NewAPIServiceClient(conn), 1, emptyAccount)

	_, err = distribute.GetBookkeeping(c, 34368, 24, "io12mgttmfa2ffn9uqvn0yn37f4nz43d248l2ga85")
	if err != nil {
		log.Printf("querh bookkeeping error: %v\n", err)
	}
}
