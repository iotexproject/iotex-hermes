package dao

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/iotexproject/iotex-hermes/cmd/key"
)

type DropRecord struct {
	gorm.Model

	EndEpoch     uint64
	DelegateName string `gorm:"type:varchar(100)"`
	Voter        string `gorm:"type:varchar(41)"`
	Index        uint64
	Amount       string `gorm:"type:varchar(50)"`
	Status       string `gorm:"type:varchar(15);index:idx_drop_records_status"`
	Hash         string `gorm:"type:varchar(64)"`
	Signature    string `gorm:"type:text"`
	ErrorMessage string `gorm:"type:text"`
}

// TableName table name of DropRecord
func (DropRecord) TableName() string {
	return "drop_records"
}

// Save insert or update drop record
func (t DropRecord) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db
	}

	if t.Signature == "" {
		signature, err := key.Sign(fmt.Sprintf("%s,%s,%s", t.DelegateName, t.Amount, t.Status), privateKey)
		if err != nil {
			return err
		}
		t.Signature = signature
	}

	if t.ID == 0 {
		var count uint64
		err := tx.Model(&DropRecord{}).Where("`end_epoch` = ? and `delegate_name` = ? and `voter` = ?", t.EndEpoch, t.DelegateName, t.Voter).Count(&count).Error
		if err != nil {
			return err
		}
		if count > 0 {
			return nil
		}

		return tx.Create(&t).Error
	}
	return tx.Save(&t).Error
}

// Verify verify signature
func (t *DropRecord) Verify() error {
	return key.Verify(fmt.Sprintf("%s,%s,%s", t.DelegateName, t.Amount, t.Status), t.Signature, publicKey)
}

// FindDropRecordByLimit find by limit
func FindNewDropRecordByLimit(limit int32) (result []DropRecord, err error) {
	err = db.Limit(limit).Where("status = ?", "new").Find(&result).Error
	return
}
