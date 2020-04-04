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

	"github.com/iotexproject/iotex-hermes/util"
)

const (
	// HermesABI defines the ABI of Hermes contract
	HermesABI = `[
    {
      "constant": true,
      "inputs": [
        {
          "name": "",
          "type": "bytes32"
        }
      ],
      "name": "distributedAmount",
      "outputs": [
        {
          "name": "",
          "type": "uint256"
        }
      ],
      "payable": false,
      "stateMutability": "view",
      "type": "function"
    },
    {
      "constant": true,
      "inputs": [],
      "name": "forwardRegistration",
      "outputs": [
        {
          "name": "",
          "type": "address"
        }
      ],
      "payable": false,
      "stateMutability": "view",
      "type": "function"
    },
    {
      "constant": true,
      "inputs": [],
      "name": "multisender",
      "outputs": [
        {
          "name": "",
          "type": "address"
        }
      ],
      "payable": false,
      "stateMutability": "view",
      "type": "function"
    },
    {
      "constant": false,
      "inputs": [
        {
          "name": "addrs",
          "type": "address[]"
        }
      ],
      "name": "removeAddressesFromWhitelist",
      "outputs": [
        {
          "name": "success",
          "type": "bool"
        }
      ],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "constant": false,
      "inputs": [
        {
          "name": "addr",
          "type": "address"
        }
      ],
      "name": "removeAddressFromWhitelist",
      "outputs": [
        {
          "name": "success",
          "type": "bool"
        }
      ],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "constant": true,
      "inputs": [],
      "name": "analyticsEndpoint",
      "outputs": [
        {
          "name": "",
          "type": "string"
        }
      ],
      "payable": false,
      "stateMutability": "view",
      "type": "function"
    },
    {
      "constant": true,
      "inputs": [
        {
          "name": "",
          "type": "uint256"
        }
      ],
      "name": "endEpochs",
      "outputs": [
        {
          "name": "",
          "type": "uint256"
        }
      ],
      "payable": false,
      "stateMutability": "view",
      "type": "function"
    },
    {
      "constant": false,
      "inputs": [
        {
          "name": "addr",
          "type": "address"
        }
      ],
      "name": "addAddressToWhitelist",
      "outputs": [
        {
          "name": "success",
          "type": "bool"
        }
      ],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "constant": true,
      "inputs": [
        {
          "name": "",
          "type": "bytes32"
        }
      ],
      "name": "distributedCount",
      "outputs": [
        {
          "name": "",
          "type": "uint256"
        }
      ],
      "payable": false,
      "stateMutability": "view",
      "type": "function"
    },
    {
      "constant": true,
      "inputs": [],
      "name": "owner",
      "outputs": [
        {
          "name": "",
          "type": "address"
        }
      ],
      "payable": false,
      "stateMutability": "view",
      "type": "function"
    },
    {
      "constant": true,
      "inputs": [
        {
          "name": "",
          "type": "address"
        }
      ],
      "name": "whitelist",
      "outputs": [
        {
          "name": "",
          "type": "bool"
        }
      ],
      "payable": false,
      "stateMutability": "view",
      "type": "function"
    },
    {
      "constant": true,
      "inputs": [],
      "name": "contractStartEpoch",
      "outputs": [
        {
          "name": "",
          "type": "uint256"
        }
      ],
      "payable": false,
      "stateMutability": "view",
      "type": "function"
    },
    {
      "constant": true,
      "inputs": [
        {
          "name": "",
          "type": "bytes32"
        },
        {
          "name": "",
          "type": "uint256"
        }
      ],
      "name": "distributions",
      "outputs": [
        {
          "name": "distributedCount",
          "type": "uint256"
        },
        {
          "name": "amount",
          "type": "uint256"
        }
      ],
      "payable": false,
      "stateMutability": "view",
      "type": "function"
    },
    {
      "constant": true,
      "inputs": [
        {
          "name": "",
          "type": "bytes32"
        },
        {
          "name": "",
          "type": "address"
        }
      ],
      "name": "recipientEpochTracker",
      "outputs": [
        {
          "name": "",
          "type": "uint256"
        }
      ],
      "payable": false,
      "stateMutability": "view",
      "type": "function"
    },
    {
      "constant": false,
      "inputs": [
        {
          "name": "addrs",
          "type": "address[]"
        }
      ],
      "name": "addAddressesToWhitelist",
      "outputs": [
        {
          "name": "success",
          "type": "bool"
        }
      ],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "constant": false,
      "inputs": [
        {
          "name": "newOwner",
          "type": "address"
        }
      ],
      "name": "transferOwnership",
      "outputs": [],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "inputs": [
        {
          "name": "_contractStartEpoch",
          "type": "uint256"
        },
        {
          "name": "_multisendAddress",
          "type": "address"
        },
        {
          "name": "_forwardRegistrationAddress",
          "type": "address"
        },
        {
          "name": "_analyticsEndpoint",
          "type": "string"
        }
      ],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "constructor"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": false,
          "name": "startEpoch",
          "type": "uint256"
        },
        {
          "indexed": false,
          "name": "endEpoch",
          "type": "uint256"
        },
        {
          "indexed": true,
          "name": "delegateName",
          "type": "bytes32"
        },
        {
          "indexed": false,
          "name": "numOfRecipients",
          "type": "uint256"
        },
        {
          "indexed": false,
          "name": "totalAmount",
          "type": "uint256"
        }
      ],
      "name": "Distribute",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": false,
          "name": "endEpoch",
          "type": "uint256"
        },
        {
          "indexed": false,
          "name": "delegateNames",
          "type": "bytes32[]"
        }
      ],
      "name": "CommitDistributions",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": false,
          "name": "addr",
          "type": "address"
        }
      ],
      "name": "WhitelistedAddressAdded",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": false,
          "name": "addr",
          "type": "address"
        }
      ],
      "name": "WhitelistedAddressRemoved",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": true,
          "name": "previousOwner",
          "type": "address"
        },
        {
          "indexed": true,
          "name": "newOwner",
          "type": "address"
        }
      ],
      "name": "OwnershipTransferred",
      "type": "event"
    },
    {
      "constant": true,
      "inputs": [],
      "name": "getEndEpochCount",
      "outputs": [
        {
          "name": "",
          "type": "uint256"
        }
      ],
      "payable": false,
      "stateMutability": "view",
      "type": "function"
    },
    {
      "constant": false,
      "inputs": [
        {
          "name": "_multisendAddress",
          "type": "address"
        }
      ],
      "name": "setMultisendAddress",
      "outputs": [],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "constant": false,
      "inputs": [
        {
          "name": "_endpoint",
          "type": "string"
        }
      ],
      "name": "setAnalyticsEndpoint",
      "outputs": [],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "constant": false,
      "inputs": [
        {
          "name": "delegateName",
          "type": "bytes32"
        },
        {
          "name": "endEpoch",
          "type": "uint256"
        },
        {
          "name": "recipients",
          "type": "address[]"
        },
        {
          "name": "amounts",
          "type": "uint256[]"
        }
      ],
      "name": "distributeRewards",
      "outputs": [],
      "payable": true,
      "stateMutability": "payable",
      "type": "function"
    },
    {
      "constant": false,
      "inputs": [
        {
          "name": "endEpoch",
          "type": "uint256"
        },
        {
          "name": "delegateNames",
          "type": "bytes32[]"
        }
      ],
      "name": "commitDistributions",
      "outputs": [],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    }
  ]`

	// MultisendABI defines the ABI of multisend contract
	MultisendABI = `[
	{
		"constant": false,
		"inputs": [
			{
				"name": "recipients",
				"type": "address[]"
			},
			{
				"name": "amounts",
				"type": "uint256[]"
			},
			{
				"name": "payload",
				"type": "string"
			}
		],
		"name": "sendCoin",
		"outputs": [],
		"payable": true,
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "_limit",
				"type": "uint256"
			}
		],
		"name": "setLimit",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [],
		"name": "withdraw",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "owner",
		"outputs": [
			{
				"name": "",
				"type": "address"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "_minTips",
				"type": "uint256"
			}
		],
		"name": "setMinTips",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "limit",
		"outputs": [
			{
				"name": "",
				"type": "uint256"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "token",
				"type": "address"
			},
			{
				"name": "recipients",
				"type": "address[]"
			},
			{
				"name": "amounts",
				"type": "uint256[]"
			},
			{
				"name": "payload",
				"type": "string"
			}
		],
		"name": "sendToken",
		"outputs": [],
		"payable": true,
		"stateMutability": "payable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "minTips",
		"outputs": [
			{
				"name": "",
				"type": "uint256"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "newOwner",
				"type": "address"
			}
		],
		"name": "transferOwnership",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"name": "_minTips",
				"type": "uint256"
			},
			{
				"name": "_limit",
				"type": "uint256"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "constructor"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"name": "from",
				"type": "address"
			},
			{
				"indexed": true,
				"name": "to",
				"type": "address"
			},
			{
				"indexed": false,
				"name": "value",
				"type": "uint256"
			}
		],
		"name": "Transfer",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"name": "_token",
				"type": "address"
			},
			{
				"indexed": false,
				"name": "_totalAmount",
				"type": "uint256"
			},
			{
				"indexed": false,
				"name": "_tips",
				"type": "uint256"
			},
			{
				"indexed": false,
				"name": "_payload",
				"type": "string"
			}
		],
		"name": "Receipt",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"name": "_owner",
				"type": "address"
			},
			{
				"indexed": false,
				"name": "_balance",
				"type": "uint256"
			}
		],
		"name": "Withdraw",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"name": "previousOwner",
				"type": "address"
			},
			{
				"indexed": true,
				"name": "newOwner",
				"type": "address"
			}
		],
		"name": "OwnershipTransferred",
		"type": "event"
	}]`
)

var DistributeCmd = &cobra.Command{
	Use:   "distribute DELEGATE",
	Short: "Distribute rewards for delegate",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return distributeReward()
	},
}

type DistributionInfo struct {
	DelegateName  string
	RecipientList []common.Address
	AmountList    []*big.Int
}

func distributeReward() error {
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
				return errors.New("invalid distributed count")
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

	lastEndEpoch, err := getLastEndEpoch(c)
	if err != nil {
		return nil, nil, nil, err
	}
	startEpoch := lastEndEpoch + 1

	epochIntervalStr := util.MustFetchNonEmptyParam("EPOCH_INTERVAL")
	epochInterval, err := strconv.Atoi(epochIntervalStr)
	if err != nil {
		return nil, nil, nil, err
	}
	endEpoch := lastEndEpoch + uint64(epochInterval)
	if lastEndEpoch == uint64(0) {
		startEpoch, err = getContractStartEpoch(c)
		if err != nil {
			return nil, nil, nil, err
		}
		firstEndEpochStr := util.MustFetchNonEmptyParam("FIRST_END_EPOCH")
		firstEndEpoch, err := strconv.Atoi(firstEndEpochStr)
		if err != nil {
			return nil, nil, nil, err
		}
		endEpoch = uint64(firstEndEpoch)
	}
	// get the current epoch on the chain and validate the endEpoch
	resp, err := c.API().GetChainMeta(context.Background(), &iotexapi.GetChainMetaRequest{})
	if err != nil {
		return nil, nil, nil, err
	}
	curEpoch := resp.ChainMeta.Epoch.Num
	if endEpoch >= curEpoch {
		return nil, nil, nil, errors.New("end epoch is not ready")
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

	totalAmount := new(big.Int).Set(minTips)
	for _, amount := range amountList {
		totalAmount.Add(totalAmount, amount)
	}
	fmt.Printf("Delegate Name: %s, Group Total Voter Count: %d, Group Total Amount: %s\n", delegateName,
		len(voterAddrList), totalAmount.String())

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

func getLastEndEpoch(c iotex.AuthedClient) (uint64, error) {
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
