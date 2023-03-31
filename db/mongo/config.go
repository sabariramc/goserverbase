package mongo

type Config struct {
	ConnectionString  string
	DatabaseName      string
	MinConnectionPool uint64
	MaxConnectionPool uint64
}
