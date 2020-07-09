package distribute

import (
	"context"
	"encoding/hex"
	"math/big"
	"strconv"
	"time"

	"github.com/iotexproject/go-pkgs/hash"
	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/pkg/errors"

	"github.com/iotexproject/iotex-hermes/util"
)

// AddDeposit add deposit to bucket
func AddDeposit(
	c iotex.AuthedClient,
	bucketID uint64,
	amount *big.Int,
) (hash.Hash256, error) {
	ctx := context.Background()

	gasPriceStr := util.MustFetchNonEmptyParam("GAS_PRICE")
	gasPrice, ok := big.NewInt(0).SetString(gasPriceStr, 10)
	if !ok {
		return hash.ZeroHash256, errors.New("failed to convert string to big int")
	}
	gasLimitStr := util.MustFetchNonEmptyParam("GAS_LIMIT")
	gasLimit, err := strconv.Atoi(gasLimitStr)
	if err != nil {
		return hash.ZeroHash256, err
	}
	h, err := c.Staking().AddDeposit(bucketID, amount).SetGasPrice(gasPrice).SetGasLimit(uint64(gasLimit)).Call(ctx)
	if err != nil {
		return hash.ZeroHash256, err
	}
	sleepIntervalStr := util.MustFetchNonEmptyParam("SLEEP_INTERVAL")
	sleepInterval, err := strconv.Atoi(sleepIntervalStr)
	if err != nil {
		return hash.ZeroHash256, err
	}
	time.Sleep(time.Duration(sleepInterval) * time.Second)

	resp, err := c.API().GetReceiptByAction(ctx, &iotexapi.GetReceiptByActionRequest{
		ActionHash: hex.EncodeToString(h[:]),
	})
	if err != nil {
		return hash.ZeroHash256, err
	}
	if resp.ReceiptInfo.Receipt.Status != 1 {
		return hash.ZeroHash256, errors.Errorf("add deposit staking failed: %x", h)
	}
	return h, nil
}
