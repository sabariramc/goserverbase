package csfle

// Config holds the configuration settings for Client-Side Field Level Encryption (CSFLE).
type Config struct {
	CryptSharedLibPath string                            // Path to the shared library for encryption.
	KeyVaultNamespace  string                            // Namespace for the key vault in MongoDB.
	SchemaMap          map[string]interface{}            // Schema map for defining encryption rules.
	KMSCredentials     map[string]map[string]interface{} // Credentials for Key Management Services (KMS).
}
