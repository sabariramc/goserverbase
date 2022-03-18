package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type clientCache map[string]*clientCacheData

type clientCacheData struct {
	client     *mongo.Client
	expireTime time.Time
}

func (c clientCache) cache(key string, client *mongo.Client) {
	c[key] = &clientCacheData{expireTime: time.Now().Add(time.Minute * 10), client: client}
}
