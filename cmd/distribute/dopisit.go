package distribute

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-hermes/util"
	"math/big"
	"strings"
)

func GetBucketID(c iotex.AuthedClient, voter common.Address) (int64, error) {
	cstring := util.MustFetchNonEmptyParam("AUTO_DEPOSIT_CONTRACT_ADDRESS")
	caddr, err := address.FromString(cstring)
	if err != nil {
		return 0, err
	}
	autoDepositABI, err := abi.JSON(strings.NewReader(AutoDepositABI))
	if err != nil {
		return 0, err
	}

	data, err := c.Contract(caddr, autoDepositABI).Read("bucket", voter).Call(context.Background())
	if err != nil {
		return 0, err
	}
	var bucketID *big.Int
	if err := data.Unmarshal(&bucketID); err != nil {
		return 0, err
	}
	return bucketID.Int64(), nil
}
