package distribute

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iotexproject/go-pkgs/hash"
	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-antenna-go/v2/account"
	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"github.com/pkg/errors"

	"github.com/iotexproject/iotex-hermes/cmd/dao"
	"github.com/iotexproject/iotex-hermes/util"
)

// GetBucketID query bucketID from contract
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

// Sender send drop record
type Sender struct {
	Accounts []account.Account
}

type accountSender struct {
	account   account.Account
	records   []dao.DropRecord
	waitGroup *sync.WaitGroup
}

func (s *accountSender) send() {
	endpoint := util.MustFetchNonEmptyParam("IO_ENDPOINT")
	conn, err := iotex.NewDefaultGRPCConn(endpoint)
	if err != nil {
		log.Fatalf("create grpc error: %v", err)
	}
	defer conn.Close()
	client := iotex.NewAuthedClient(iotexapi.NewAPIServiceClient(conn), s.account)

	for _, record := range s.records {
		if record.Verify() != nil {
			record.Status = "error_signature"
			err = record.Save(dao.DB())
			if err != nil {
				log.Fatalf("save drop records error: %v", err)
			}
			continue
		}
		amount, ok := big.NewInt(0).SetString(record.Amount, 10)
		if !ok {
			log.Printf("can't convert staking amount: %v\n", record.Amount)
		}
		h, err := addDeposit(client, record.ID, record.Index, amount)
		if err != nil {
			log.Printf("add deposit %d error: %v\n", record.ID, err)
			record.Status = "error"
			record.ErrorMessage = err.Error()
			err = record.Save(dao.DB())
			if err != nil {
				log.Fatalf("save error drop records %d:%s error: %v", record.ID, record.Voter, err)
			}
			continue
		}
		record.Hash = hex.EncodeToString(h[:])
		record.Signature = ""
		record.Status = "completed"
		err = record.Save(dao.DB())
		if err != nil {
			log.Fatalf("save success drop records %d:%s error: %v", record.ID, record.Voter, err)
		}
	}

	s.records = nil
	if s.waitGroup != nil {
		s.waitGroup.Done()
		s.waitGroup = nil
	}
}

func addDeposit(
	c iotex.AuthedClient,
	recordID uint,
	bucketID uint64,
	amount *big.Int,
) (hash.Hash256, error) {
	ctx := context.Background()

	gasPriceStr := util.MustFetchNonEmptyParam("GAS_PRICE")
	gasPrice, ok := big.NewInt(0).SetString(gasPriceStr, 10)
	if !ok {
		return hash.ZeroHash256, errors.New("failed to convert string to big int")
	}
	gasLimit := 10000

	gas := big.NewInt(0).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	if amount.Cmp(gas) <= 0 {
		log.Printf("amount %s less than gas for %d\n", amount.String(), recordID)
		return hash.ZeroHash256, nil
	}

	h, err := c.Staking().AddDeposit(bucketID, big.NewInt(0).Sub(amount, gas)).SetGasPrice(gasPrice).SetGasLimit(uint64(gasLimit)).Call(ctx)
	if err != nil {
		return hash.ZeroHash256, err
	}
	time.Sleep(5 * time.Second)

	for i := 0; i < 30; i++ {
		resp, err := c.API().GetReceiptByAction(ctx, &iotexapi.GetReceiptByActionRequest{
			ActionHash: hex.EncodeToString(h[:]),
		})
		if err != nil {
			if strings.Contains(err.Error(), "code = NotFound") {
				time.Sleep(1 * time.Second)
				continue
			}
			return hash.ZeroHash256, err
		}
		if resp.ReceiptInfo.Receipt.Status != 1 {
			return hash.ZeroHash256, errors.Errorf("add deposit staking failed: %x", h)
		}
		return h, nil
	}
	return hash.ZeroHash256, errors.Errorf("add deposit error by exhausted retry, index=%d, hash: %x", bucketID, h)
}

// Send send records
func (s *Sender) Send() {
	fmt.Println("Begin add deposit to bucket")
	for {
		records, err := dao.FindNewDropRecordByLimit(10000)
		if err != nil {
			log.Fatalf("query drop records error: %v", err)
		}
		if len(records) == 0 {
			break
		}

		shard := len(s.Accounts)
		if len(records) < shard || shard == 1 {
			sender := &accountSender{
				account: s.Accounts[0],
				records: records,
			}
			sender.send()
		} else {
			wg := sync.WaitGroup{}
			wg.Add(shard)
			size := len(records) / shard
			for i := 0; i < shard; i++ {
				end := size * (i + 1)
				if i == shard-1 {
					end = len(records)
				}
				sender := &accountSender{
					account:   s.Accounts[i],
					records:   records[i*size : end],
					waitGroup: &wg,
				}
				go sender.send()
			}
			wg.Wait()
		}
	}
	fmt.Println("Add deposit to bucket successful.")
}

// NewSender new sender instance
func NewSender() (*Sender, error) {
	pwd := util.MustFetchNonEmptyParam("VAULT_PASSWORD")
	acc, err := util.GetVaultAccount(pwd)
	if err != nil {
		return nil, err
	}
	// verify the account matches the reward address
	if acc.Address().String() != util.MustFetchNonEmptyParam("VAULT_ADDRESS") {
		return nil, fmt.Errorf("key and address do not match")
	}

	return &Sender{
		Accounts: []account.Account{acc},
	}, nil
}
