package config

type MongoCSFLEConfig struct {
	KeyVaultNamespace string
	MasterKeyARN      string
}

type ServerConfig struct {
	Host        string
	Port        string
	ServiceName string
	Debug       bool
}

type RuntimeConfig struct {
	GoMaxProcs int
}
