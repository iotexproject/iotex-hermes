// Copyright (c) 2020 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package claim

import (
	"math/big"
	"os"
	"testing"

	"github.com/iotexproject/iotex-antenna-go/v2/account"
	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/stretchr/testify/require"
)

const (
	ioEndpoint     = "api.testnet.iotex.one:443"
	testPrivateKey = "2394db684d2d14586e16ec597ce9222a2e552265a58da2a9218a09e3ccff8893"
	sleepInterval  = "20"
)

func TestClaimReward(t *testing.T) {
	require := require.New(t)

	account, err := account.HexStringToAccount(testPrivateKey)
	require.NoError(err)

	conn, err := iotex.NewDefaultGRPCConn(ioEndpoint)
	require.NoError(err)
	defer conn.Close()

	c := iotex.NewAuthedClient(iotexapi.NewAPIServiceClient(conn), 1, account)

	unClaimedBalance, err := getUnclaimedBalance(c)
	require.NoError(err)
	require.True(unClaimedBalance.Sign() > 0)

	os.Setenv("SLEEP_INTERVAL", sleepInterval)
	require.NoError(claim(c, big.NewInt(1)))
}
