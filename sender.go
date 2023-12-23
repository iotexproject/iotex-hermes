// Copyright (c) 2020 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package main

import (
	"log"

	"github.com/iotexproject/iotex-hermes/cmd/dao"
	"github.com/iotexproject/iotex-hermes/cmd/distribute"
	"github.com/iotexproject/iotex-hermes/util"
)

func main() {
	err := dao.ConnectDatabase()
	if err != nil {
		log.Fatalf("create database error: %v\n", err)
	}

	notifier, err := distribute.NewNotifier(util.MustFetchNonEmptyParam("LARK_ENDPOINT"), util.MustFetchNonEmptyParam("LARK_KEY"))
	if err != nil {
		log.Fatalf("new notifier error: %v\n", err)
	}

	sender, err := distribute.NewSender(notifier)
	if err != nil {
		log.Fatalf("new notifier error: %v\n", err)
	}
	go sender.Send()

	forever := make(chan bool)
	<-forever
}
