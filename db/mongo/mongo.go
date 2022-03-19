package mongo

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

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
	log     *log.Logger
}

var ErrNoDocuments = mongo.ErrNoDocuments

var mongoCache clientCache = make(clientCache)

func GetDefaultClient(ctx context.Context, logger *log.Logger, mongoURI string) (*Mongo, error) {
	return GetClient(ctx, logger, mongoURI)
}

func GetClient(ctx context.Context, logger *log.Logger, mongoURI string) (*Mongo, error) {
	uriHash := utils.GetHash(mongoURI)
	cachedClient, ok := mongoCache[uriHash]
	var client *mongo.Client
	var err error
	if ok && time.Now().Before(cachedClient.expireTime) {
		logger.Debug(ctx, "Mongo client taken from cache", nil)
		client = cachedClient.client
	} else {
		client, err = NewMongoClient(ctx, logger, mongoURI)
		if err != nil {
			return nil, err
		}
		mongoCache.cache(uriHash, client)
	}
	return &Mongo{client: client, log: logger, urlHash: uriHash}, nil
}

func NewMongoClient(ctx context.Context, logger *log.Logger, mongoURI string) (*mongo.Client, error) {
	connectionOptions := options.Client()
	connectionOptions.ApplyURI(mongoURI)
	connectionOptions.SetConnectTimeout(time.Minute)
	connectionOptions.SetMaxConnIdleTime(time.Minute * 12)
	client, err := mongo.Connect(ctx, connectionOptions)
	if err != nil {
		logger.Error(ctx, "Error creating mongo connection", err)
		return nil, err
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error(ctx, "Error pinging mongo server", err)
		return nil, err
	}
	return client, nil
}

func (m *Mongo) GetDataKey(ctx context.Context, keyVaultNamespace, keyAltName string, provider MasterKeyProvider) (res string, err error) {

	// configuring encryption options by setting the keyVault namespace and the kms providers information
	// we configure this client to fetch the master key so that we can
	// create a data key in the next step

	// look for a data key
	keyVault := strings.Split(keyVaultNamespace, ".")
	db := keyVault[0]
	coll := keyVault[1]
	var dataKey bson.M
	err = m.client.
		Database(db).
		Collection(coll).
		FindOne(ctx, bson.M{"keyAltNames": keyAltName}).
		Decode(&dataKey)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			m.log.Error(ctx, "error encountered while attempting to find key", err)
			return
		}
		var data *string
		data, err = m.CreateDataKey(ctx, keyVaultNamespace, keyAltName, provider)
		if err != nil {
			return
		}
		res = *data

	} else {
		res = base64.StdEncoding.EncodeToString(dataKey["_id"].(primitive.Binary).Data)
	}
	return

}

func (m *Mongo) CreateDataKey(ctx context.Context, keyVaultNamespace, keyAltName string, provider MasterKeyProvider) (*string, error) {
	// specify the master key information that will be used to
	// encrypt the data key(s) that will in turn be used to encrypt
	// fields, and create the data key
	clientEncryptionOpts := options.ClientEncryption().SetKeyVaultNamespace(keyVaultNamespace).SetKmsProviders(provider.Credentials())
	clientEnc, err := mongo.NewClientEncryption(m.client, clientEncryptionOpts)
	if err != nil {
		m.log.Error(ctx, "Error creating encryption client", err)
		return nil, err
	}
	dataKeyOpts := options.DataKey().
		SetMasterKey(provider.DataKeyOpts()).
		SetKeyAltNames([]string{keyAltName})
	dataKeyID, err := clientEnc.CreateDataKey(ctx, provider.Name(), dataKeyOpts)
	if err != nil {
		m.log.Error(ctx, "Error creating data key", err)
		return nil, err
	}
	res := base64.StdEncoding.EncodeToString(dataKeyID.Data)
	return &res, nil
}
