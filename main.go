// Copyright (c) 2020 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package main

import (
	"context"
	"log"
	"time"

	"github.com/iotexproject/iotex-antenna-go/v2/account"
	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"

	"github.com/iotexproject/iotex-hermes/cmd/claim"
	"github.com/iotexproject/iotex-hermes/cmd/dao"
	"github.com/iotexproject/iotex-hermes/cmd/distribute"
	"github.com/iotexproject/iotex-hermes/util"
)

// main runs the hermes command
func main() {
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
	c := iotex.NewAuthedClient(iotexapi.NewAPIServiceClient(conn), emptyAccount)

	err = dao.ConnectDatabase()
	if err != nil {
		log.Fatalf("create database error: %v\n", err)
	}

	for {
		lastEndEpoch, err := distribute.GetLastEndEpoch(c)
		if err != nil {
			log.Fatalf("get last end epoch error: %v\n", err)
		}
		startEpoch := lastEndEpoch + 1

		resp, err := c.API().GetChainMeta(context.Background(), &iotexapi.GetChainMetaRequest{})
		if err != nil {
			log.Fatalf("get chain meta error: %v\n", err)
		}
		curEpoch := resp.ChainMeta.Epoch.Num

		endEpoch := startEpoch + 23

		if endEpoch+2 > curEpoch {
			duration := time.Duration(endEpoch + 2 - curEpoch)
			log.Printf("waiting %d hours for next distribute", duration)
			time.Sleep(duration * time.Hour)
			continue
		}
		err = claim.Reward()
		if err != nil {
			log.Fatalf("claim reward error: %v\n", err)
		}
		err = distribute.Reward()
		if err != nil {
			log.Fatalf("distribute reward error: %v\n", err)
		}

		sender, err := distribute.NewSender()
		if err != nil {
			log.Fatalf("new sender error: %v\n", err)
		}
		sender.Send()
	}
}
