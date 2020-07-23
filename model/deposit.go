package model

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotexproject/iotex-hermes/key"
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
	Signature  string `gorm:"type:varchar(64)"`
}

// CreateJob Create a Job use some params
func CreateJob(startEpoch, endEpoch *big.Int, delegateNameList []string, addressList []common.Address, bucketIDList []uint64, amountList []*big.Int) (*Job, error) {
	if len(addressList) != len(bucketIDList) || len(addressList) != len(amountList) || len(addressList) != len(delegateNameList) {
		return nil, fmt.Errorf("DelegateNameList or addressList or bucketIDList or amountList are not equal in length")
	}

	deposits := make([]*Deposit, len(delegateNameList))
	for i := range delegateNameList {
		deposits[i] = &Deposit{
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

	return job, job.create(deposits)
}

func (j *Job) create(deposits []*Deposit) error {
	if j == nil {
		return fmt.Errorf("Job object is empty.")
	}

	if err := loadKeys(); err != nil {
		return err
	}

	signature, err := key.Sign(fmt.Sprintf("%s,%s,%d", j.StartEpoch, j.EndEpoch, STATUS_OPEN), privateKey)
	if err != nil {
		return err
	}
	j.Signature = signature
	j.Status = STATUS_OPEN
	dbTran := Begin()
	if err := dbTran.Create(j).Error; err != nil {
		dbTran.Rollback()
		return err
	}

	for i := range deposits {
		signature, err := key.Sign(fmt.Sprintf("%d,%s,%d", deposits[i].VoterBucketID, deposits[i].Amount, STATUS_OPEN), privateKey)
		if err != nil {
			return err
		}
		deposits[i].Signature = signature
		deposits[i].JobID = j.ID
		deposits[i].Status = STATUS_OPEN
		if err := dbTran.Create(deposits[i]).Error; err != nil {
			db.Rollback()
			return err
		}
	}
	return dbTran.Commit().Error
}

func (j *Job) Update() error {
	if err := loadKeys(); err != nil {
		return err
	}

	signature, err := key.Sign(fmt.Sprintf("%s,%s,%d", j.StartEpoch, j.EndEpoch, STATUS_OPEN), privateKey)
	if err != nil {
		return err
	}

	return DB().Model(j).Update(map[string]interface{}{"status": STATUS_COMPLATE, "signature": signature}).Error
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

	if err := loadKeys(); err != nil {
		return nil, err
	}

	if err := key.Verify(fmt.Sprintf("%s,%s,%d", job.StartEpoch, job.EndEpoch, job.Status), job.Signature, publicKey); err != nil {
		return nil, err
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

	if err := loadKeys(); err != nil {
		return err
	}

	for i := range deposits {
		if err := key.Verify(fmt.Sprintf("%d,%s,%d", deposits[i].VoterBucketID, deposits[i].Amount, deposits[i].Status), deposits[i].Signature, publicKey); err != nil {
			return err
		}
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

func (j *Job) IsComplate() (bool, error) {
	if err := loadKeys(); err != nil {
		return false, fmt.Errorf("can't load keys err: %v", err)
	}

	if err := key.Verify(fmt.Sprintf("%s,%s,%d", j.StartEpoch, j.EndEpoch, j.Status), j.Signature, publicKey); err != nil {
		return false, fmt.Errorf("verification fails for %s, err: %v", fmt.Sprintf("%s,%s,%d", j.StartEpoch, j.EndEpoch, j.Status), err)
	}

	return j.Status == STATUS_COMPLATE, nil
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
	Signature     string `gorm:"type:varchar(64)"`
}

func (d *Deposit) Update() error {
	if err := loadKeys(); err != nil {
		return err
	}

	signature, err := key.Sign(fmt.Sprintf("%d,%s,%d", d.VoterBucketID, d.Amount, STATUS_COMPLATE), privateKey)
	if err != nil {
		return err
	}

	return DB().Model(d).Update(map[string]interface{}{"status": STATUS_COMPLATE, "signature": signature}).Error
}

func (d *Deposit) IsComplate() (bool, error) {
	if err := loadKeys(); err != nil {
		return false, fmt.Errorf("can't load keys err: %v", err)
	}

	if err := key.Verify(fmt.Sprintf("%d,%s,%d", d.VoterBucketID, d.Amount, d.Status), d.Signature, publicKey); err != nil {
		return false, fmt.Errorf("verification fails for %s, err: %v", fmt.Sprintf("%d,%s,%d", d.VoterBucketID, d.Amount, d.Status), err)
	}

	return d.Status == STATUS_COMPLATE, nil
}
