// Copyright (c) 2020 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package distribute

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
	"github.com/spf13/cobra"

	"github.com/iotexproject/iotex-hermes/cmd/dao"
	"github.com/iotexproject/iotex-hermes/util"
)

// DistributeCmd is the distribute command
var DistributeCmd = &cobra.Command{
	Use:   "distribute DELEGATE",
	Short: "Distribute rewards for delegate",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return Reward()
	},
}

// DistributionInfo defines the distribution information
type DistributionInfo struct {
	DelegateName  string
	RecipientList []common.Address
	AmountList    []*big.Int
}

// Reward distribute reward to voter group by delegate
func Reward() error {
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

	// query GraphQL to get the distribution list
	endEpoch, tip, distributions, err := getDistribution(c)
	if err != nil {
		return err
	}

	// call distribution contract to send out rewards
	chunkSizeStr := util.MustFetchNonEmptyParam("CHUNK_SIZE")
	chunkSize, err := strconv.Atoi(chunkSizeStr)
	if err != nil {
		return err
	}
	delegateNames := make([][32]byte, 0, len(distributions))
	for _, dist := range distributions {
		delegateNames = append(delegateNames, stringToBytes32(dist.DelegateName))
		divAddrList, divAmountList, err := splitRecipients(chunkSize, dist.RecipientList, dist.AmountList)
		if err != nil {
			return err
		}
		for {
			distrbutedCount, err := getDistributedCount(c, dist.DelegateName)
			if err != nil {
				return err
			}
			// distribution is done for the delegate
			if int(distrbutedCount) == len(dist.RecipientList) {
				break
			}
			if int(distrbutedCount)%chunkSize != 0 {
				return fmt.Errorf("invalid distributed count, Delegate Name: %s, Distributed Count: %d, Number of Recipients: %d",
					dist.DelegateName, distrbutedCount, len(dist.RecipientList))
			}
			nextGroup := int(distrbutedCount) / chunkSize
			if err := sendRewards(c, dist.DelegateName, endEpoch, tip, divAddrList[nextGroup], divAmountList[nextGroup]); err != nil {
				return err
			}
		}
	}
	return commitDistributions(c, endEpoch, delegateNames)
}

func getDistribution(c iotex.AuthedClient) (*big.Int, *big.Int, []*DistributionInfo, error) {
	minTips, err := getMinTips(c)
	if err != nil {
		return nil, nil, nil, err
	}

	lastEndEpoch, err := GetLastEndEpoch(c)
	if err != nil {
		return nil, nil, nil, err
	}
	startEpoch := lastEndEpoch + 1

	resp, err := c.API().GetChainMeta(context.Background(), &iotexapi.GetChainMetaRequest{})
	if err != nil {
		return nil, nil, nil, err
	}
	curEpoch := resp.ChainMeta.Epoch.Num

	endEpoch := startEpoch + 23

	if endEpoch+2 > curEpoch {
		return nil, nil, nil, fmt.Errorf("invalid end epoch, Current Epoch: %d, End Epoch: %d",
			curEpoch, endEpoch)
	}

	if startEpoch > endEpoch {
		return nil, nil, nil, fmt.Errorf("invalid epoch range, Start Epoch: %d, End Epoch: %d",
			startEpoch, endEpoch)
	}

	fmt.Printf("Distribution Start Epoch: %d\n", startEpoch)
	fmt.Printf("Distribution End Epoch: %d\n", endEpoch)

	rewardAddress := c.Account().Address().String()
	epochCount := endEpoch - startEpoch + 1
	distributions, err := getBookkeeping(startEpoch, epochCount, rewardAddress)
	if err != nil {
		return nil, nil, nil, err
	}
	return big.NewInt(int64(endEpoch)), minTips, distributions, nil
}

func sendRewards(
	c iotex.AuthedClient,
	delegateName string,
	endEpoch *big.Int,
	minTips *big.Int,
	voterAddrList []common.Address,
	amountList []*big.Int,
) error {
	cstring := util.MustFetchNonEmptyParam("HERMES_CONTRACT_ADDRESS")
	caddr, err := address.FromString(cstring)
	if err != nil {
		return err
	}

	// call distribution contract to send out rewards
	ctx := context.Background()
	hermesABI, err := abi.JSON(strings.NewReader(HermesABI))
	if err != nil {
		return err
	}

	for i := 0; i < len(voterAddrList); i++ {
		bucketID, err := GetBucketID(c, voterAddrList[i])
		if err != nil {
			fmt.Printf("Query bucketID from contract error: %v\n", err)
			continue
		}
		if bucketID != -1 {
			addr, err := address.FromBytes(voterAddrList[i][:])
			if err != nil {
				fmt.Printf("Convert address error: %v\n", err)
				continue
			}
			drop := dao.DropRecord{
				EndEpoch:     endEpoch.Uint64(),
				DelegateName: delegateName,
				Voter:        addr.String(),
				Amount:       amountList[i].String(),
				Index:        uint64(bucketID),
				Status:       "new",
			}
			err = drop.Save(dao.DB())
			if err != nil {
				fmt.Printf("Save drop record error: %v\n", err)
				continue
			}
			amountList[i] = big.NewInt(0)
		}
	}

	totalAmount := new(big.Int).Set(minTips)
	for _, amount := range amountList {
		totalAmount.Add(totalAmount, amount)
	}
	fmt.Printf("Delegate Name: %s, Group Total Voter Count: %d, Group Total Amount: %s, Tip: %s\n", delegateName,
		len(voterAddrList), totalAmount.String(), minTips.String())

	name := stringToBytes32(delegateName)

	gasPriceStr := util.MustFetchNonEmptyParam("GAS_PRICE")
	gasPrice, ok := big.NewInt(0).SetString(gasPriceStr, 10)
	if !ok {
		return errors.New("failed to convert string to big int")
	}
	gasLimitStr := util.MustFetchNonEmptyParam("GAS_LIMIT")
	gasLimit, err := strconv.Atoi(gasLimitStr)
	if err != nil {
		return err
	}
	h, err := c.Contract(caddr, hermesABI).Execute("distributeRewards", name, endEpoch, voterAddrList, amountList).
		SetAmount(totalAmount).SetGasPrice(gasPrice).SetGasLimit(uint64(gasLimit)).Call(ctx)
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
		ActionHash: hex.EncodeToString(h[:]),
	})
	if err != nil {
		return err
	}
	if resp.ReceiptInfo.Receipt.Status != 1 {
		return errors.Errorf("distributeRewards failed: %x", h)
	}
	return nil
}

func commitDistributions(c iotex.AuthedClient, endEpoch *big.Int, delegateNames [][32]byte) error {
	cstring := util.MustFetchNonEmptyParam("HERMES_CONTRACT_ADDRESS")
	caddr, err := address.FromString(cstring)
	if err != nil {
		return err
	}

	// call distribution contract to send out rewards
	ctx := context.Background()
	hermesABI, err := abi.JSON(strings.NewReader(HermesABI))
	if err != nil {
		return err
	}

	gasPriceStr := util.MustFetchNonEmptyParam("GAS_PRICE")
	gasPrice, ok := big.NewInt(0).SetString(gasPriceStr, 10)
	if !ok {
		return errors.New("failed to convert string to big int")
	}
	gasLimitStr := util.MustFetchNonEmptyParam("GAS_LIMIT")
	gasLimit, err := strconv.Atoi(gasLimitStr)
	if err != nil {
		return err
	}
	h, err := c.Contract(caddr, hermesABI).Execute("commitDistributions", endEpoch, delegateNames).
		SetGasPrice(gasPrice).SetGasLimit(uint64(gasLimit)).Call(ctx)
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
		ActionHash: hex.EncodeToString(h[:]),
	})
	if err != nil {
		return err
	}
	if resp.ReceiptInfo.Receipt.Status != 1 {
		return errors.Errorf("commitDistributions failed: %x", h)
	}

	fmt.Println("successfully distribute rewards")
	return nil
}

func getMinTips(c iotex.AuthedClient) (*big.Int, error) {
	cstring := util.MustFetchNonEmptyParam("MULTISEND_CONTRACT_ADDRESS")
	caddr, err := address.FromString(cstring)
	if err != nil {
		return nil, err
	}
	multisendABI, err := abi.JSON(strings.NewReader(MultisendABI))
	if err != nil {
		return nil, err
	}
	data, err := c.Contract(caddr, multisendABI).Read("minTips").Call(context.Background())
	if err != nil {
		return nil, err
	}
	var minTips *big.Int
	if err := data.Unmarshal(&minTips); err != nil {
		return nil, err
	}

	fmt.Printf("MultiSend Contract: %s, min tip: %s\n", cstring, minTips.String())
	return minTips, nil
}

func getContractStartEpoch(c iotex.AuthedClient) (uint64, error) {
	cstring := util.MustFetchNonEmptyParam("HERMES_CONTRACT_ADDRESS")
	caddr, err := address.FromString(cstring)
	if err != nil {
		return 0, err
	}
	hermesABI, err := abi.JSON(strings.NewReader(HermesABI))
	if err != nil {
		return 0, err
	}
	data, err := c.Contract(caddr, hermesABI).Read("contractStartEpoch").Call(context.Background())
	if err != nil {
		return 0, err
	}
	var contractStartEpoch *big.Int
	if err := data.Unmarshal(&contractStartEpoch); err != nil {
		return 0, err
	}
	return contractStartEpoch.Uint64(), nil
}

// GetLastEndEpoch get last end epoch from hermes contract
func GetLastEndEpoch(c iotex.AuthedClient) (uint64, error) {
	cstring := util.MustFetchNonEmptyParam("HERMES_CONTRACT_ADDRESS")
	caddr, err := address.FromString(cstring)
	if err != nil {
		return 0, err
	}
	hermesABI, err := abi.JSON(strings.NewReader(HermesABI))
	if err != nil {
		return 0, err
	}
	data, err := c.Contract(caddr, hermesABI).Read("getEndEpochCount").Call(context.Background())
	if err != nil {
		return 0, err
	}
	var endEpochCount *big.Int
	if err := data.Unmarshal(&endEpochCount); err != nil {
		return 0, err
	}

	if endEpochCount.String() == "0" {
		return 0, nil
	}
	data, err = c.Contract(caddr, hermesABI).Read("endEpochs", endEpochCount.Sub(endEpochCount, big.NewInt(1))).Call(context.Background())
	if err != nil {
		return 0, err
	}
	var lastEndEpoch *big.Int
	if err := data.Unmarshal(&lastEndEpoch); err != nil {
		return 0, err
	}
	return lastEndEpoch.Uint64(), nil
}

func getDistributedCount(c iotex.AuthedClient, delegateName string) (uint64, error) {
	cstring := util.MustFetchNonEmptyParam("HERMES_CONTRACT_ADDRESS")
	caddr, err := address.FromString(cstring)
	if err != nil {
		return 0, err
	}
	hermesABI, err := abi.JSON(strings.NewReader(HermesABI))
	if err != nil {
		return 0, err
	}

	name := stringToBytes32(delegateName)
	data, err := c.Contract(caddr, hermesABI).Read("distributedCount", name).Call(context.Background())
	if err != nil {
		return 0, err
	}
	var distributedCount *big.Int
	if err := data.Unmarshal(&distributedCount); err != nil {
		return 0, err
	}
	return distributedCount.Uint64(), nil
}

func getBookkeeping(startEpoch uint64, epochCount uint64, rewardAddress string) ([]*DistributionInfo, error) {
	type query struct {
		Hermes struct {
			Exist              graphql.Boolean
			HermesDistribution []struct {
				DelegateName       graphql.String
				RewardDistribution []struct {
					VoterIotexAddress graphql.String
					Amount            graphql.String
				}
				StakingIotexAddress graphql.String
				VoterCount          graphql.Int
				WaiveServiceFee     graphql.Boolean
				Refund              graphql.String
			}
		} `graphql:"hermes(startEpoch: $startEpoch, epochCount: $epochCount, rewardAddress: $rewardAddress, waiverThreshold: $waiverThreshold)"`
	}

	analyticsEndpoint := util.MustFetchNonEmptyParam("ANALYTICS_ENDPOINT")

	gqlClient := graphql.NewClient(analyticsEndpoint, nil)

	waiverThresholdStr := util.MustFetchNonEmptyParam("WAIVER_THRESHOLD")
	waiverThreshold, err := strconv.Atoi(waiverThresholdStr)
	if err != nil {
		return nil, err
	}

	// make sure every epoch does not miss hermes info
	for epoch := startEpoch; epoch < startEpoch+epochCount; epoch++ {
		tempVariables := map[string]interface{}{
			"startEpoch":      graphql.Int(epoch),
			"epochCount":      graphql.Int(1),
			"rewardAddress":   graphql.String(rewardAddress),
			"waiverThreshold": graphql.Int(waiverThreshold),
		}
		var tempOutput query
		if err := gqlClient.Query(context.Background(), &tempOutput, tempVariables); err != nil {
			return nil, err
		}
		if !tempOutput.Hermes.Exist {
			return nil, errors.New(fmt.Sprintf("bookkeeping info doesn't exist for Epoch %d\n", epoch))
		}
	}

	variables := map[string]interface{}{
		"startEpoch":      graphql.Int(startEpoch),
		"epochCount":      graphql.Int(epochCount),
		"rewardAddress":   graphql.String(rewardAddress),
		"waiverThreshold": graphql.Int(waiverThreshold),
	}
	var output query
	if err := gqlClient.Query(context.Background(), &output, variables); err != nil {
		return nil, err
	}

	if !output.Hermes.Exist {
		return nil, errors.New("bookkeeping info doesn't exist within the epoch range")
	}

	distributions := make([]*DistributionInfo, 0, len(output.Hermes.HermesDistribution))
	for _, hermesDistribution := range output.Hermes.HermesDistribution {
		distributionMap := make(map[string]*big.Int)
		for _, rewardDistribution := range hermesDistribution.RewardDistribution {
			amount, ok := big.NewInt(0).SetString(string(rewardDistribution.Amount), 10)
			if !ok {
				return nil, errors.New("failed to convert string to big int")
			}
			distributionMap[string(rewardDistribution.VoterIotexAddress)] = amount
		}
		// Add delegate to the map
		refund, ok := big.NewInt(0).SetString(string(hermesDistribution.Refund), 10)
		if !ok {
			return nil, errors.New("failed to convert string to big int")
		}
		// charge fees
		serviceFee := big.NewInt(0)
		if !hermesDistribution.WaiveServiceFee {
			if serviceFee, refund, err = calculateServiceFee(int64(hermesDistribution.VoterCount), refund); err != nil {
				return nil, err
			}
		}
		fmt.Printf("Delegate Name: %s, Service Fee: %s, Refund: %s\n", string(hermesDistribution.DelegateName),
			serviceFee.String(), refund.String())

		delegateIotexStakingAddr := string(hermesDistribution.StakingIotexAddress)
		if _, ok := distributionMap[delegateIotexStakingAddr]; !ok {
			distributionMap[delegateIotexStakingAddr] = refund
		} else {
			distributionMap[delegateIotexStakingAddr].Add(distributionMap[delegateIotexStakingAddr], refund)
		}

		var keys []string
		for k := range distributionMap {
			keys = append(keys, k)
		}
		// sort recipient addresses
		sort.Strings(keys)

		recipientAddrList := make([]common.Address, 0, len(distributionMap))
		amountList := make([]*big.Int, 0, len(distributionMap))
		for _, k := range keys {
			caddr, err := ioAddrToEvmAddr(k)
			if err != nil {
				return nil, err
			}
			recipientAddrList = append(recipientAddrList, caddr)
			amountList = append(amountList, distributionMap[k])
		}

		distributions = append(distributions, &DistributionInfo{
			DelegateName:  string(hermesDistribution.DelegateName),
			RecipientList: recipientAddrList,
			AmountList:    amountList,
		})
	}
	// sort distributions by delegate name
	sort.Slice(distributions, func(i, j int) bool { return distributions[i].DelegateName < distributions[j].DelegateName })

	return distributions, nil
}

func calculateServiceFee(voterCount int64, refund *big.Int) (*big.Int, *big.Int, error) {
	baseChargeStr := util.MustFetchNonEmptyParam("BASE_CHARGE")
	baseCharge, ok := big.NewInt(0).SetString(baseChargeStr, 10)
	if !ok {
		return nil, nil, errors.New("failed to convert string to big int")
	}
	chargePerRecipientStr := util.MustFetchNonEmptyParam("CHARGE_PER_RECIPIENT")
	chargePerRecipient, ok := big.NewInt(0).SetString(chargePerRecipientStr, 10)
	if !ok {
		return nil, nil, errors.New("failed to convert string to big int")
	}
	serviceFee := baseCharge
	extraCharge := big.NewInt(voterCount)
	extraCharge.Mul(extraCharge, chargePerRecipient)
	serviceFee.Add(serviceFee, extraCharge)
	balance := new(big.Int).Set(refund)
	refund.Sub(refund, serviceFee)
	if refund.Sign() < 0 {
		refund = big.NewInt(0)
		serviceFee = balance
	}
	return serviceFee, refund, nil
}

func splitRecipients(chunkSize int, recipientAddrList []common.Address, amountList []*big.Int) ([][]common.Address, [][]*big.Int, error) {
	if len(recipientAddrList) != len(amountList) {
		return nil, nil, errors.New("length does not match")
	}
	var divAddrList [][]common.Address
	var divAmountList [][]*big.Int

	for i := 0; i < len(recipientAddrList); i += chunkSize {
		end := i + chunkSize

		if end > len(recipientAddrList) {
			end = len(recipientAddrList)
		}

		divAddrList = append(divAddrList, recipientAddrList[i:end])
		divAmountList = append(divAmountList, amountList[i:end])
	}

	return divAddrList, divAmountList, nil
}

// ioAddrToEvmAddr converts IoTeX address into evm address
func ioAddrToEvmAddr(ioAddr string) (common.Address, error) {
	// temporary fix
	if ioAddr == "io16y9wk2xnwurvtgmd2mds2gcdfe2lmzad6dcw29" {
		ioAddr = "io16dkdajys8609qxf78wmmzssgfgvqkk0funzp0r"
	}
	address, err := address.FromString(ioAddr)
	if err != nil {
		return common.Address{}, err
	}
	return common.BytesToAddress(address.Bytes()), nil
}

// stringToBytes32 converts string to bytes32
func stringToBytes32(delegateName string) [32]byte {
	var name [32]byte
	copy(name[:], delegateName)
	return name
}
