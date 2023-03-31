package config

type MySqlConnectionConfig struct {
	Host         string
	Port         string
	DatabaseName string
	Username     string
	Password     string
	Timezone     string
	Charset      string
}

type MongoConfig struct {
	ConnectionString  string
	DatabaseName      string
	MinConnectionPool uint64
	MaxConnectionPool uint64
}

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
