package mysql

import (
	"time"

	"gorm.io/gorm"
)

type BaseMysqlModel struct {
	ID        uint           `json:"-" gorm:"primarykey"`
	CreatedAt time.Time      `json:"createdAt" gorm:"not null"`
	UpdatedAt time.Time      `json:"updatedAt" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	CreatedBy string         `json:"createdBy" gorm:"not null;type:varchar(36)"`
	UpdatedBy string         `json:"updatedBy" gorm:"not null;type:varchar(36)"`
}
