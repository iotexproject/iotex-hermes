package distribute

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
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
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

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

	bucketID, err := data.Unmarshal()
	if err != nil {
		return 0, err
	}
	return bucketID[0].(*big.Int).Int64(), nil
}

// Sender send drop record
type Sender struct {
	Accounts []account.Account
	Notifier *Notifier
}

type accountSender struct {
	account   account.Account
	records   []dao.DropRecord
	waitGroup *sync.WaitGroup
	notifier  *Notifier
}

type analyserData struct {
	EpochNumber  uint64 `json:"epochNumber"`
	DelegateName string `json:"delegateName"`
	VoterAddress string `json:"voterAddress"`
	ActHash      string `json:"actHash"`
	BucketID     uint64 `json:"bucketID"`
	Amount       string `json:"amount"`
}

var bucketStateMap = make(map[uint64]bool)

func (s *accountSender) send() {
	tls := util.MustFetchNonEmptyParam("RPC_TLS")
	endpoint := util.MustFetchNonEmptyParam("IO_ENDPOINT")

	var conn *grpc.ClientConn
	var err error

	if tls == "true" {
		conn, err = iotex.NewDefaultGRPCConn(endpoint)
		if err != nil {
			log.Fatalf("create grpc error: %v", err)
		}
	} else {
		conn, err = iotex.NewGRPCConnWithoutTLS(endpoint)
		if err != nil {
			log.Fatalf("create grpc error: %v", err)
		}
	}
	defer conn.Close()
	client := iotex.NewAuthedClient(iotexapi.NewAPIServiceClient(conn), 1, s.account)

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
		h, ignore, ra, err := addDepositOrTransfer(client, record.ID, record.Index, record.Voter, record.DelegateName, amount)
		if err != nil {
			if ignore {
				log.Printf("add deposit %d error: %v\n", record.ID, err)
				time.Sleep(1 * time.Minute)
				continue
			} else {
				if !strings.HasSuffix(err.Error(), "insufficient funds for gas * price + value") {
					s.notifier.SendMessage(fmt.Sprintf("Deposit %d error: %v", record.ID, err))
					break
				}
				if !strings.HasPrefix(err.Error(), "add deposit error by exhausted retry") {
					s.notifier.SendMessage(fmt.Sprintf("Deposit %d error: %v", record.ID, err))
				}
				record.Status = "error"
				record.ErrorMessage = err.Error()
				err = record.Save(dao.DB())
				if err != nil {
					log.Fatalf("save error drop records %d:%s error: %v", record.ID, record.Voter, err)
				}
			}
		}
		record.Hash = hex.EncodeToString(h[:])
		record.Signature = ""
		record.Status = "completed"
		err = record.Save(dao.DB())
		if err != nil {
			log.Fatalf("save success drop records %d:%s error: %v", record.ID, record.Voter, err)
		}

		ad := analyserData{
			EpochNumber:  record.EndEpoch,
			DelegateName: record.DelegateName,
			VoterAddress: record.Voter,
			BucketID:     record.Index,
			ActHash:      record.Hash,
			Amount:       ra.String(),
		}
		postAnalyserData(&ad)
	}

	s.records = nil
	if s.waitGroup != nil {
		s.waitGroup.Done()
		s.waitGroup = nil
	}
}

func postAnalyserData(ad *analyserData) {
	data, err := json.Marshal(ad)
	if err != nil {
		log.Printf("marshal data error: %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		request, err := http.NewRequestWithContext(
			ctx,
			"POST",
			"https://analyser-api.iotex.io/api.HermesService.HermesDropRecords",
			bytes.NewBuffer(data),
		)
		if err != nil {
			log.Printf("new request error: %v\n", err)
		}
		request.Header.Set("Content-Type", "application/json; charset=UTF-8")

		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			log.Printf("post data error: %v\n", err)
			return
		}
		defer response.Body.Close()
		io.Copy(ioutil.Discard, request.Body)
	}(ctx)

	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Duration(11 * time.Second)):
		fmt.Println("post data timeout")
		return
	}
}

func checkAutoStake(c iotex.AuthedClient, bucketID uint64) (bool, error) {
	state, ok := bucketStateMap[bucketID]
	if ok {
		return state, nil
	}

	method := &iotexapi.ReadStakingDataMethod{
		Method: iotexapi.ReadStakingDataMethod_BUCKETS_BY_INDEXES,
	}
	methodBytes, err := proto.Marshal(method)
	if err != nil {
		return false, err
	}
	arguments := &iotexapi.ReadStakingDataRequest{
		Request: &iotexapi.ReadStakingDataRequest_BucketsByIndexes{
			BucketsByIndexes: &iotexapi.ReadStakingDataRequest_VoteBucketsByIndexes{
				Index: []uint64{bucketID},
			},
		},
	}
	argumentsBytes, err := proto.Marshal(arguments)
	if err != nil {
		return false, err
	}

	res, err := c.API().ReadState(context.Background(), &iotexapi.ReadStateRequest{
		ProtocolID: []byte("staking"),
		MethodName: methodBytes,
		Arguments:  [][]byte{argumentsBytes},
		Height:     "",
	})
	if err != nil {
		return false, err
	}
	var result iotextypes.VoteBucketList
	err = proto.Unmarshal(res.Data, &result)
	if err != nil {
		return false, err
	}
	if len(result.Buckets) == 0 {
		return false, fmt.Errorf("can't find bucket %d", bucketID)
	}
	bucketStateMap[bucketID] = result.Buckets[0].AutoStake
	return result.Buckets[0].AutoStake, nil
}

func addDepositOrTransfer(
	c iotex.AuthedClient,
	recordID uint,
	bucketID uint64,
	voter string,
	delegateName string,
	amount *big.Int,
) (hash.Hash256, bool, *big.Int, error) {
	ctx := context.Background()

	gasPriceStr := util.MustFetchNonEmptyParam("GAS_PRICE")
	gasPrice, ok := big.NewInt(0).SetString(gasPriceStr, 10)
	if !ok {
		return hash.ZeroHash256, true, nil, errors.New("failed to convert string to big int")
	}
	// TODO change to 10000?
	gasLimit := 13000

	gas := big.NewInt(0).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	if amount.Cmp(gas) <= 0 {
		log.Printf("amount %s less than gas for %d\n", amount.String(), recordID)
		return hash.ZeroHash256, true, nil, nil
	}

	autoStake, err := checkAutoStake(c, bucketID)
	if err != nil {
		log.Printf("check auto stake bucket error: %v", err)
	}

	var h hash.Hash256
	ra := big.NewInt(0).Sub(amount, gas)
	if !autoStake {
		to, _ := address.FromString(voter)
		h, err = c.Transfer(to, ra).SetGasPrice(gasPrice).SetGasLimit(uint64(gasLimit)).Call(ctx)
	} else {
		h, err = c.Staking().AddDeposit(bucketID, ra).SetGasPrice(gasPrice).SetGasLimit(uint64(gasLimit)).Call(ctx)
	}

	if err != nil {
		return hash.ZeroHash256, true, nil, err
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
			return h, false, nil, err
		}
		if resp.ReceiptInfo.Receipt.Status == 204 {
			delete(bucketStateMap, bucketID)
			return addDepositOrTransfer(c, recordID, bucketID, voter, delegateName, amount)
		}
		if resp.ReceiptInfo.Receipt.Status != 1 {
			return h, false, nil, errors.Errorf("add deposit staking failed: %x", h)
		}
		return h, false, ra, nil
	}
	return h, false, nil, errors.Errorf("add deposit error by exhausted retry, index=%d, hash: %x", bucketID, h)
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
			time.Sleep(5 * time.Minute)
			continue
		}
		s.Notifier.SendMessage(fmt.Sprintf("Begin send %d compound hermes rewards", len(records)))

		shard := len(s.Accounts)
		if len(records) < shard || shard == 1 {
			sender := &accountSender{
				account:  s.Accounts[0],
				records:  records,
				notifier: s.Notifier,
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
					notifier:  s.Notifier,
				}
				go sender.send()
			}
			wg.Wait()
		}
	}
}

// NewSender new sender instance
func NewSender(notifier *Notifier) (*Sender, error) {
	pwd := util.MustFetchNonEmptyParam("VAULT_PASSWORD")
	acc, err := util.GetAccount("./sender/.sender.json", pwd)
	if err != nil {
		return nil, err
	}

	return &Sender{
		Accounts: []account.Account{acc},
		Notifier: notifier,
	}, nil
}
