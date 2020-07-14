package model

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const (
	STATUS_OPEN = uint8(iota)
	STATUS_COMPLATE
)

// Job defines a job structrue
type Job struct {
	ID         uint64 `gorm:"primary_key"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
	StartEpoch string `gorm:"type:varchar(64)"`
	EndEpoch   string `gorm:"type:varchar(64)"`
	Status     uint8  `gorm:"typr:TINYINT(1)"`
	Deposits   []Deposit
}

// CreateJob Create a Job use some params
func CreateJob(startEpoch, endEpoch *big.Int, delegateNameList []string, addressList []common.Address, bucketIDList []uint64, amountList []*big.Int) (*Job, error) {
	if len(addressList) != len(bucketIDList) || len(addressList) != len(amountList) || len(addressList) != len(delegateNameList) {
		return nil, fmt.Errorf("DelegateNameList or addressList or bucketIDList or amountList are not equal in length")
	}

	depositMs := make([]*Deposit, len(delegateNameList))
	for i := range delegateNameList {
		depositMs[i] = &Deposit{
			DelegateName:  delegateNameList[i],
			VoterAddress:  addressList[i].String(),
			VoterBucketID: bucketIDList[i],
			Amount:        amountList[i].String(),
		}
	}

	job := &Job{
		StartEpoch: startEpoch.String(),
		EndEpoch:   endEpoch.String(),
	}

	return job, job.create(depositMs)
}

func (j *Job) create(deposits []*Deposit) error {
	if j == nil {
		return fmt.Errorf("Job object is empty.")
	}
	j.Status = STATUS_OPEN
	dbTran := Begin()
	if err := dbTran.Create(j).Error; err != nil {
		dbTran.Rollback()
		return err
	}

	for i := range deposits {
		deposits[i].JobID = j.ID
		deposits[i].Status = STATUS_OPEN
		if err := dbTran.Create(deposits[i]).Error; err != nil {
			db.Rollback()
			return err
		}
	}
	dbTran.Commit()
	return nil
}

func (j *Job) Update() error {
	return DB().Model(j).Update("status", STATUS_COMPLATE).Error
}

func GetLatest() (*Job, error) {
	var job *Job
	result := DB().Last(job)
	if result.Error != nil {
		if result.RecordNotFound() {
			return nil, nil
		}
		return nil, result.Error
	}
	return job, nil
}

func (j *Job) LoadDeposits() error {
	if j.ID == 0 {
		return fmt.Errorf("Can't load an empty job.")
	}
	var deposits []Deposit
	if err := DB().Model(j).Related(&deposits).Error; err != nil {
		return err
	}

	j.Deposits = deposits
	return nil
}

func (j *Job) Delete() error {
	for i := range j.Deposits {
		DB().Delete(j.Deposits[i])
	}
	return DB().Delete(j).Error
}

func (j *Job) IsComplate() bool {
	return j.Status == STATUS_COMPLATE
}

// Deposit defines a Deposit structrue
type Deposit struct {
	ID            uint64 `gorm:"primary_key"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	DelegateName  string `gorm:"type:varchar(64)"`
	VoterAddress  string `gorm:"type:varchar(64)"`
	VoterBucketID uint64
	Amount        string `gorm:"type:varchar(64)"`
	Status        uint8  `gorm:"typr:TINYINT(1)"`
	JobID         uint64
}

func (d *Deposit) Update() error {
	return DB().Model(d).Update("status", STATUS_COMPLATE).Error
}

func (d *Deposit) IsComplate() bool {
	return d.Status == STATUS_COMPLATE
}
