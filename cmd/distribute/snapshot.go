package distribute

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

type Snapshot struct {
	DivAddrList     [][]common.Address `json:"divAddrList"`
	DivAmountList   [][]*big.Int       `json:"divAmountList"`
	TotalRecipients int                `json:"totalRecipients"`
}

func (s *Snapshot) Save(name string, epoch uint64) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("./snapshots/%s-%d.json", name, epoch)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file %s exist", path)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

func LoadSnapshot(name string, epoch uint64) (*Snapshot, error) {
	path := fmt.Sprintf("./snapshots/%s-%d.json", name, epoch)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var snapshot Snapshot
	err = json.Unmarshal(data, &snapshot)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}
