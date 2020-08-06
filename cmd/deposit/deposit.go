// Copyright (c) 2020 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package deposit

import (
	"fmt"
	"math/big"

	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/spf13/cobra"

	"github.com/iotexproject/iotex-hermes/model"
	"github.com/iotexproject/iotex-hermes/util"
)

// DistributeCmd is the distribute command
var DepositCmd = &cobra.Command{
	Use:   "deposit DELEGATE",
	Short: "Deposit to bucket for voter",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return depositToBucket()
	},
}

func depositToBucket() error {
	// Get latest job
	latestJob, err := model.GetLatest()
	if err != nil {
		return err
	}

	// If is already completed. return
	if i, e := latestJob.IsCompleted(); e != nil {
		return e
	} else if i {
		return nil
	}

	// Load deposits of the job.
	if err := latestJob.LoadDeposits(); err != nil {
		return err
	}

	pwd := util.MustFetchNonEmptyParam("VAULT_PASSWORD")
	account, err := util.GetVaultAccount(pwd)
	if err != nil {
		return err
	}
	// verify the account matches the reward address
	if account.Address().String() != util.MustFetchNonEmptyParam("VAULT_ADDRESS") {
		return fmt.Errorf("key and address do not match")
	}

	endpoint := util.MustFetchNonEmptyParam("IO_ENDPOINT")
	conn, err := iotex.NewDefaultGRPCConn(endpoint)
	if err != nil {
		return err
	}
	defer conn.Close()
	c := iotex.NewAuthedClient(iotexapi.NewAPIServiceClient(conn), account)

	for i := range latestJob.Deposits {
		if ic, e := latestJob.Deposits[i].IsCompleted(); e != nil {
			return e
		} else if ic {
			// If this voter is completed, continue next.
			continue
		}

		if latestJob.Deposits[i].Validate() != nil {
			err = latestJob.Deposits[i].UpdateStatus(model.StatusError)
			if err != nil {
				return err
			}
			continue
		}

		amount, ok := big.NewInt(0).SetString(latestJob.Deposits[i].Amount, 10)
		if !ok {
			return fmt.Errorf("amount %s can not convert a big.Int", latestJob.Deposits[i].Amount)
		}
		_, err := AddDeposit(c, latestJob.Deposits[i].VoterBucketID, amount, latestJob.Deposits[i].DelegateName)
		if err != nil {
			return err
		}
		if err := latestJob.Deposits[i].UpdateStatus(model.StatusCompleted); err != nil {
			return err
		}
	}

	// The loop did not let the function exit,
	// indicating that all deposit have been completed, set job to completed
	return latestJob.Update()
}
