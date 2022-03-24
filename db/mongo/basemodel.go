package mongo

import (
	"time"
)

type BaseMongoDocument struct {
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	CreatedBy string    `json:"createdBy" bson:"createdBy"`
	UpdatedBy string    `json:"updatedBy" bson:"updatedBy"`
}

type BaseMongoModel struct {
	BaseMongoDocument `bson:",inline"`
}
