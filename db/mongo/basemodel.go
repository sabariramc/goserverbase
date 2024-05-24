package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BaseMongoDocument defines a basic structure for a MongoDB document.
type BaseMongoDocument struct {
	CreatedAt *time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"` // Time when the document was created.
	UpdatedAt *time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"` // Time when the document was last updated.
	CreatedBy *string    `json:"createdBy,omitempty" bson:"createdBy,omitempty"` // Identifier of the user who created the document.
	UpdatedBy *string    `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"` // Identifier of the user who last updated the document.
}

// SetCreateParam sets the creation parameters for the document.
func (b *BaseMongoDocument) SetCreateParam(actionBy string) {
	actionAt := time.Now()
	b.CreatedAt = &actionAt
	b.UpdatedAt = &actionAt
	b.CreatedBy = &actionBy
	b.UpdatedBy = &actionBy
}

// SetUpdateParam sets the update parameters for the document.
func (b *BaseMongoDocument) SetUpdateParam(actionBy string) {
	actionAt := time.Now()
	b.UpdatedAt = &actionAt
	b.UpdatedBy = &actionBy
}

// BaseMongoModel extends BaseMongoDocument with an ObjectId.
type BaseMongoModel struct {
	ID                 *primitive.ObjectID `json:"-" bson:"_id,omitempty"` // Unique identifier of the MongoDB document.
	*BaseMongoDocument `bson:",inline"`    // Embedded BaseMongoDocument.
}

// SetCreateParam sets the creation parameters for the model.
func (b *BaseMongoModel) SetCreateParam(actionBy string) {
	if b.BaseMongoDocument == nil {
		b.BaseMongoDocument = &BaseMongoDocument{}
	}
	b.BaseMongoDocument.SetCreateParam(actionBy)
}

// SetUpdateParam sets the update parameters for the model.
func (b *BaseMongoModel) SetUpdateParam(actionBy string) {
	if b.BaseMongoDocument == nil {
		b.BaseMongoDocument = &BaseMongoDocument{}
	}
	b.BaseMongoDocument.SetUpdateParam(actionBy)
}
