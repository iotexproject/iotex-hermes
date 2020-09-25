package dao

import "github.com/jinzhu/gorm"

type DropRecord struct {
	gorm.Model

	EndEpoch     uint64
	delegateName string `gorm:"type:varchar(100)"`
	voter        string `gorm:"type:varchar(41)"`
	Amount       string `gorm:"type:varchar(50)"`
	Status       string `gorm:"type:varchar(15);index:idx_drop_records_status"`
	Signature    string `gorm:"type:text"`
	ErrorMessage string `gorm:"type:text"`
}

// TableName table name of DropRecord
func (DropRecord) TableName() string {
	return "drop_records"
}
