package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BaseMongoDocument struct {
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	CreatedBy string    `json:"createdBy" bson:"createdBy"`
	UpdatedBy string    `json:"updatedBy" bson:"updatedBy"`
}

type BaseMongoModel struct {
	ID                *primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	BaseMongoDocument `bson:",inline"`
}
