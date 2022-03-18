package mongo

type MasterKeyProvider interface {
	Name() string
	Credentials() map[string]map[string]interface{}
	DataKeyOpts() interface{}
}
