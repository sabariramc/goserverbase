package mysql

import (
	"time"
)

type BaseMysqlModel struct {
	ID        uint       `json:"-" gorm:"primarykey"`
	CreatedAt *time.Time `json:"createdAt,omitempty" gorm:"not null"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty" gorm:"not null"`
	CreatedBy *string    `json:"createdBy,omitempty" gorm:"not null;type:varchar(36)"`
	UpdatedBy *string    `json:"updatedBy,omitempty" gorm:"not null;type:varchar(36)"`
}

func (b *BaseMysqlModel) SetCreateParam(actionBy string) {
	actionAt := time.Now()
	b.CreatedAt = &actionAt
	b.UpdatedAt = &actionAt
	b.CreatedBy = &actionBy
	b.UpdatedBy = &actionBy
}

func (b *BaseMysqlModel) SetUpdateParam(actionBy string) {
	actionAt := time.Now()
	b.UpdatedAt = &actionAt
	b.UpdatedBy = &actionBy
}
