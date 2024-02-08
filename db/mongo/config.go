package mongo

type Config struct {
	AppName           string
	ConnectionString  string
	MinConnectionPool uint64
	MaxConnectionPool uint64
	EnableLog         bool
}
