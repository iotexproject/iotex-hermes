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
	"strconv"
	"time"

	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/iotexproject/iotex-proto/golang/protocol"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

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

	endpoint := util.MustFetchNonEmptyParam("IO_ENDPOINT")
	conn, err := iotex.NewDefaultGRPCConn(endpoint)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := iotex.NewAuthedClient(iotexapi.NewAPIServiceClient(conn), account)

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
	sleepIntervalStr := util.MustFetchNonEmptyParam("SLEEP_INTERVAL")
	sleepInterval, err := strconv.Atoi(sleepIntervalStr)
	if err != nil {
		return err
	}
	time.Sleep(time.Duration(sleepInterval) * time.Second)

	resp, err := c.API().GetReceiptByAction(ctx, &iotexapi.GetReceiptByActionRequest{
		ActionHash: hex.EncodeToString(hash[:]),
	})
	if err != nil {
		return err
	}
	if resp.ReceiptInfo.Receipt.Status != 1 {
		return errors.Errorf("claim rewards failed: %x", hash)
	}
	fmt.Println("successfully claim rewards")
	return nil
}
