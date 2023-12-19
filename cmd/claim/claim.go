// Copyright (c) 2020 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package claim

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/iotexproject/go-pkgs/hash"
	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/iotexproject/iotex-proto/golang/protocol"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/iotexproject/iotex-hermes/util"
)

// ClaimCmd is the claim command
var ClaimCmd = &cobra.Command{
	Use:   "claim DELEGATE",
	Short: "Claim rewards for delegate",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		_, err := Reward()
		return err
	},
}

// Reward is claim reward from contract
func Reward() (*big.Int, error) {
	pwd := util.MustFetchNonEmptyParam("VAULT_PASSWORD")
	account, err := util.GetVaultAccount(pwd)
	if err != nil {
		return nil, err
	}

	tls := util.MustFetchNonEmptyParam("RPC_TLS")
	endpoint := util.MustFetchNonEmptyParam("IO_ENDPOINT")
	var conn *grpc.ClientConn

	if tls == "true" {
		conn, err = iotex.NewDefaultGRPCConn(endpoint)
		if err != nil {
			return nil, err
		}
	} else {
		conn, err = iotex.NewGRPCConnWithoutTLS(endpoint)
		if err != nil {
			return nil, err
		}
	}
	defer conn.Close()

	c := iotex.NewAuthedClient(iotexapi.NewAPIServiceClient(conn), 1, account)

	// get current epoch and block height
	resp, err := c.API().GetChainMeta(context.Background(), &iotexapi.GetChainMetaRequest{})
	if err != nil {
		return nil, err
	}
	curEpoch := resp.ChainMeta.Epoch.Num
	curHeight := resp.ChainMeta.Height

	fmt.Printf("Current Epoch Number: %d\n", curEpoch)
	fmt.Printf("Current Block Height: %d\n", curHeight)

	unclaimedBalance, err := getUnclaimedBalance(c)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Unclaimed Balance: %s\n", unclaimedBalance.String())
	err = claim(c, unclaimedBalance)
	return unclaimedBalance, err
}

func getUnclaimedBalance(c iotex.AuthedClient) (*big.Int, error) {
	request := &iotexapi.ReadStateRequest{
		ProtocolID: []byte(protocol.RewardingProtocolID),
		MethodName: []byte(protocol.ReadUnclaimedBalanceMethodName),
		Arguments:  [][]byte{[]byte(c.Account().Address().String())},
	}
	response, err := c.API().ReadState(context.Background(), request)
	if err != nil {
		return nil, err
	}
	unclaimedBlance, ok := big.NewInt(0).SetString(string(response.Data), 10)
	if !ok {
		return nil, errors.New("failed to convert string to big int")
	}
	return unclaimedBlance, nil
}

func claim(c iotex.AuthedClient, unclaimedBalance *big.Int) error {
	ctx := context.Background()
	hash, err := c.ClaimReward(unclaimedBalance).Call(ctx)
	if err != nil {
		return err
	}

	err = checkActionReceipt(c, hash)
	if err != nil {
		return err
	}
	fmt.Println("successfully claim rewards")
	return nil
}

func checkActionReceipt(c iotex.AuthedClient, hash hash.Hash256) error {
	time.Sleep(5 * time.Second)
	var resp *iotexapi.GetReceiptByActionResponse
	var err error
	for i := 0; i < 120; i++ {
		resp, err = c.API().GetReceiptByAction(context.Background(), &iotexapi.GetReceiptByActionRequest{
			ActionHash: hex.EncodeToString(hash[:]),
		})
		if err != nil {
			if strings.Contains(err.Error(), "code = NotFound") {
				time.Sleep(2 * time.Second)
				continue
			}
			return err
		}
		if resp.ReceiptInfo.Receipt.Status != 1 {
			return errors.Errorf("action %x check receipt failed", hash)
		}
		return nil
	}
	fmt.Printf("action %x check receipt not found\n", hash)
	return err
}
