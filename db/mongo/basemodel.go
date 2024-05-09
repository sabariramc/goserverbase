package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BaseMongoDocument defines a simple base struct for mongo document
type BaseMongoDocument struct {
	CreatedAt *time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	CreatedBy *string    `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	UpdatedBy *string    `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"`
}

func (b *BaseMongoDocument) SetCreateParam(actionBy string) {
	actionAt := time.Now()
	b.CreatedAt = &actionAt
	b.UpdatedAt = &actionAt
	b.CreatedBy = &actionBy
	b.UpdatedBy = &actionBy
}

func (b *BaseMongoDocument) SetUpdateParam(actionBy string) {
	actionAt := time.Now()
	b.UpdatedAt = &actionAt
	b.UpdatedBy = &actionBy
}

// BaseMongoModel extends BaseMongoDocument with the bson objectId
type BaseMongoModel struct {
	ID                 *primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	*BaseMongoDocument `bson:",inline"`
}

func (b *BaseMongoModel) SetCreateParam(actionBy string) {
	if b.BaseMongoDocument == nil {
		b.BaseMongoDocument = &BaseMongoDocument{}
	}
	b.BaseMongoDocument.SetCreateParam(actionBy)
}

func (b *BaseMongoModel) SetUpdateParam(actionBy string) {
	if b.BaseMongoDocument == nil {
		b.BaseMongoDocument = &BaseMongoDocument{}
	}
	b.BaseMongoDocument.SetUpdateParam(actionBy)
}
