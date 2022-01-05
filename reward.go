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

	"github.com/iotexproject/iotex-address/address"
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

	notifier, err := distribute.NewNotifier()
	if err != nil {
		log.Fatalf("new notifier error: %v\n", err)
	}

	retry := 0
	for {
		if retry == 3 {
			log.Fatalf("retry 2 times failure, exit.")
		}
		lastEndEpoch, err := distribute.GetLastEndEpoch(c)
		if err != nil {
			log.Printf("get last end epoch error: %v\n", err)
			retry++
			time.Sleep(5 * time.Minute)
			continue
		}
		startEpoch := lastEndEpoch + 1

		resp, err := c.API().GetChainMeta(context.Background(), &iotexapi.GetChainMetaRequest{})
		if err != nil {
			log.Printf("get chain meta error: %v\n", err)
			retry++
			time.Sleep(5 * time.Minute)
			continue
		}
		curEpoch := resp.ChainMeta.Epoch.Num

		endEpoch := startEpoch + 23

		if endEpoch+2 > curEpoch {
			resp, err := c.API().GetChainMeta(context.Background(), &iotexapi.GetChainMetaRequest{})
			if err != nil {
				log.Printf("get chain meta error: %v\n", err)
				retry++
				time.Sleep(5 * time.Minute)
				continue
			}
			curEpoch = resp.ChainMeta.Epoch.Num
			if endEpoch+2-curEpoch > 0 {
				duration := time.Duration(endEpoch + 2 - curEpoch)
				log.Printf("waiting %d hours for next distribute", duration)
				time.Sleep(duration * time.Hour)
				continue
			}
		}
		err = claim.Reward()
		if err != nil {
			log.Printf("claim reward error: %v\n", err)
			retry++
			time.Sleep(5 * time.Minute)
			continue
		}
		lastDeposit, lastEpoch, err := dao.SumByEndEpoch(lastEndEpoch)
		if err != nil {
			log.Printf("sum last deposit error: %v\n", err)
			retry++
			continue
		}

		sender, err := address.FromString(util.MustFetchNonEmptyParam("SENDER_ADDR"))
		if err != nil {
			log.Printf("get sender address error: %v\n", err)
			retry++
			continue
		}
		err = distribute.Reward(notifier, lastDeposit, lastEpoch, sender)
		if err != nil {
			log.Printf("distribute reward error: %v\n", err)
			retry++
			time.Sleep(5 * time.Minute)
			continue
		}
		retry = 0
	}
}
