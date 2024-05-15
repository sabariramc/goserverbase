package csfle

import "github.com/sabariramc/goserverbase/v6/db/mongo"

type Config struct {
	*mongo.Config
	CryptSharedLibPath string
	KeyVaultNamespace  string
	SchemaMap          map[string]interface{}
	KMSCredentials     map[string]map[string]interface{}
}
