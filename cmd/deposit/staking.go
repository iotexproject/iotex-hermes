package deposit

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iotexproject/go-pkgs/hash"
	"github.com/iotexproject/iotex-address/address"
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
	delegateName string,
) (hash.Hash256, error) {
	ctx := context.Background()

	gasPriceStr := util.MustFetchNonEmptyParam("GAS_PRICE")
	gasPrice, ok := big.NewInt(0).SetString(gasPriceStr, 10)
	if !ok {
		printError(delegateName, bucketID, amount,
			fmt.Errorf("set gasPrice faild. ENV GAS_PRICE set error, failed to convert string to big int"))
		return hash.ZeroHash256, errors.New("failed to convert string to big int")
	}
	gasLimitStr := util.MustFetchNonEmptyParam("GAS_LIMIT")
	gasLimit, err := strconv.Atoi(gasLimitStr)
	if err != nil {
		printError(delegateName, bucketID, amount,
			fmt.Errorf("set gasLimit faild. ENV GAS_LIMIT set error"))
		return hash.ZeroHash256, err
	}
	h, err := c.Staking().AddDeposit(bucketID, amount).SetGasPrice(gasPrice).SetGasLimit(uint64(gasLimit)).Call(ctx)
	if err != nil {
		printError(delegateName, bucketID, amount,
			fmt.Errorf("execute addDeposit faild: %v", err))
		return hash.ZeroHash256, err
	}
	sleepIntervalStr := util.MustFetchNonEmptyParam("SLEEP_INTERVAL")
	sleepInterval, err := strconv.Atoi(sleepIntervalStr)
	if err != nil {
		printError(delegateName, bucketID, amount,
			fmt.Errorf("set sleepInterval faild. ENV SLEEP_INTERVAL set error: %v", err))
		return hash.ZeroHash256, err
	}
	time.Sleep(time.Duration(sleepInterval) * time.Second)

	resp, err := c.API().GetReceiptByAction(ctx, &iotexapi.GetReceiptByActionRequest{
		ActionHash: hex.EncodeToString(h[:]),
	})
	if err != nil {
		printError(delegateName, bucketID, amount,
			fmt.Errorf("execute GetReceiptByAction faild: %v", err))
		return hash.ZeroHash256, err
	}
	if resp.ReceiptInfo.Receipt.Status != 1 {
		printError(delegateName, bucketID, amount,
			fmt.Errorf("get Execute Status is not 1, add deposit staking failed: %x", h))
		return hash.ZeroHash256, errors.Errorf("add deposit staking failed: %x", h)
	}

	printSuccess(delegateName, bucketID, amount)

	return h, nil
}

func printError(delegateName string, bucketID uint64, amount *big.Int, err error) {
	fmt.Printf("Delegate Name: %s, Voter Bucket ID: %d, Group Total Amount: %s. Error: %v\n", delegateName, bucketID, amount, err)
}

func printSuccess(delegateName string, bucketID uint64, amount *big.Int) {
	fmt.Printf("Delegate Name: %s, Voter Bucket ID: %d, Group Total Amount: %s\n", delegateName, bucketID, amount)
}

// GetBucketID query bucketId from contract
func GetBucketID(c iotex.AuthedClient, voter string) (int64, error) {
	cstring := util.MustFetchNonEmptyParam("AUTO_DEPOSIT_CONTRACT_ADDRESS")
	caddr, err := address.FromString(cstring)
	if err != nil {
		return 0, err
	}
	autoDepositABI, err := abi.JSON(strings.NewReader(AutoDepositABI))
	if err != nil {
		return 0, err
	}

	ioAddress, err := ioAddrToEvmAddr(voter)
	if err != nil {
		return 0, err
	}
	data, err := c.Contract(caddr, autoDepositABI).Read("bucket", ioAddress).Call(context.Background())
	if err != nil {
		return 0, err
	}
	var bucketID *big.Int
	if err := data.Unmarshal(&bucketID); err != nil {
		return 0, err
	}
	return bucketID.Int64(), nil
}

func ioAddrToEvmAddr(ioAddr string) (common.Address, error) {
	address, err := address.FromString(ioAddr)
	if err != nil {
		return common.Address{}, err
	}
	return common.BytesToAddress(address.Bytes()), nil
}
