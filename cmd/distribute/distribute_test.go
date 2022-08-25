// Copyright (c) 2020 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package distribute

import (
	"os"
	"testing"

	"github.com/iotexproject/iotex-antenna-go/v2/account"
	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/stretchr/testify/require"
)

const (
	ioEndpoint       = "api.iotex.one:443"
	multiSendAddress = "io1lvemm43lz6np0hzcqlpk0kpxxww623z5hs4mwu"
	expectedMinTips  = "5000000000000000000"
	testPrivateKey   = "a000000000000000000000000000000000000000000000000000000000000000"
)

func TestGetMinTips(t *testing.T) {
	require := require.New(t)

	account, err := account.HexStringToAccount(testPrivateKey)
	require.NoError(err)

	conn, err := iotex.NewDefaultGRPCConn(ioEndpoint)
	require.NoError(err)
	defer conn.Close()

	c := iotex.NewAuthedClient(iotexapi.NewAPIServiceClient(conn), 1, account)

	os.Setenv("MULTISEND_CONTRACT_ADDRESS", multiSendAddress)

	minTips, err := getMinTips(c)
	require.Equal(minTips.String(), expectedMinTips)
}
