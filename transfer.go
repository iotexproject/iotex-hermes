package main

import (
	"context"
	"log"
	"math/big"

	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-hermes/util"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
)

func main() {
	endpoint := util.MustFetchNonEmptyParam("IO_ENDPOINT")
	conn, err := iotex.NewDefaultGRPCConn(endpoint)
	if err != nil {
		log.Fatalf("construct grpc connection error: %v\n", err)
	}
	defer conn.Close()

	pwd := util.MustFetchNonEmptyParam("VAULT_PASSWORD")

	sender, _ := util.GetAccount("./sender/.sender.json", pwd)
	c := iotex.NewAuthedClient(iotexapi.NewAPIServiceClient(conn), sender)

	amount, _ := new(big.Int).SetString("30000000000000000000000", 10)

	receipt, _ := address.FromString("io12mgttmfa2ffn9uqvn0yn37f4nz43d248l2ga85")

	c.Transfer(receipt, amount).SetGasPrice(big.NewInt(1000000000000)).SetGasLimit(10000).Call(context.Background())
}
