package mongo

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"sabariram.com/goserverbase/log"
	"sabariram.com/goserverbase/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	urlHash string
	client  *mongo.Client
	ctx     context.Context
	log     *log.Log
}

var ErrNoDocuments = mongo.ErrNoDocuments

var mongoCache clientCache = make(clientCache)

func GetDefaultClient(ctx context.Context, mongoURI string) (*Mongo, error) {
	return GetClient(ctx, mongoURI, log.GetDefaultLogger())
}

func NewMongoClient(ctx context.Context, mongoURI string, logger *log.Log) (*mongo.Client, error) {
	connectionOptions := options.Client()
	connectionOptions.ApplyURI(mongoURI)
	connectionOptions.SetConnectTimeout(time.Minute)
	connectionOptions.SetMaxConnIdleTime(time.Minute * 12)
	client, err := mongo.Connect(ctx, connectionOptions)
	if err != nil {
		logger.Error("Error creating mongo connection", err)
		return nil, err
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error("Error pinging mongo server", err)
		return nil, err
	}
	return client, nil
}

func GetClient(ctx context.Context, mongoURI string, logger *log.Log) (*Mongo, error) {
	uriHash := utils.GetHash(mongoURI)
	cachedClient, ok := mongoCache[uriHash]
	var client *mongo.Client
	var err error
	if ok && time.Now().Before(cachedClient.expireTime) {
		logger.Debug("Mongo client taken from cache", nil)
		client = cachedClient.client
	} else {
		client, err = NewMongoClient(ctx, mongoURI, logger)
		if err != nil {
			return nil, err
		}
		mongoCache.cache(uriHash, client)
	}
	return &Mongo{client: client, ctx: ctx, log: logger, urlHash: uriHash}, nil
}

func (m *Mongo) GetDataKey(keyVaultNamespace, keyAltName string, provider MasterKeyProvider) (res string, err error) {

	// configuring encryption options by setting the keyVault namespace and the kms providers information
	// we configure this client to fetch the master key so that we can
	// create a data key in the next step
	err = xray.Capture(m.ctx, "GetMongoDataKey", func(ctx1 context.Context) error {
		clientEncryptionOpts := options.ClientEncryption().SetKeyVaultNamespace(keyVaultNamespace).SetKmsProviders(provider.Credentials())

		clientEnc, err := mongo.NewClientEncryption(m.client, clientEncryptionOpts)
		if err != nil {
			m.log.Error("Error creating encryption client", err)
			return err
		}
		// look for a data key
		keyVault := strings.Split(keyVaultNamespace, ".")
		db := keyVault[0]
		coll := keyVault[1]
		var dataKey bson.M
		err = m.client.
			Database(db).
			Collection(coll).
			FindOne(m.ctx, bson.M{"keyAltNames": keyAltName}).
			Decode(&dataKey)
		if err == mongo.ErrNoDocuments {
			// specify the master key information that will be used to
			// encrypt the data key(s) that will in turn be used to encrypt
			// fields, and create the data key
			dataKeyOpts := options.DataKey().
				SetMasterKey(provider.DataKeyOpts()).
				SetKeyAltNames([]string{keyAltName})
			dataKeyID, err := clientEnc.CreateDataKey(m.ctx, provider.Name(), dataKeyOpts)
			if err != nil {
				m.log.Error("Error creating data key", err)
				return err
			}
			res = base64.StdEncoding.EncodeToString(dataKeyID.Data)
			return nil
		}
		if err != nil {
			m.log.Error("error encountered while attempting to find key", err)
			return err
		}
		res = base64.StdEncoding.EncodeToString(dataKey["_id"].(primitive.Binary).Data)
		return nil
	})
	return

}
