package mongo

type Config struct {
	ConnectionString  string
	MinConnectionPool uint64
	MaxConnectionPool uint64
}
