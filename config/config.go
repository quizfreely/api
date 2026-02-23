package config

type Config struct {
	Port int `toml:"port"`
	DBURL string `toml:"db_url"`
	DBMigrationURL string `toml:"db_migration_url"`
	PrettyLog bool `toml:"pretty_log"`
	BasePath string `toml:"base_path"`
	EnableOAuthGoogle bool `toml:"enable_oauth_google"`
	OAuthGoogleClientID string `toml:"oauth_google_client_id"`
	OAuthGoogleClientSecret string `toml:"oauth_google_client_secret"`
	OAuthGoogleCallbackURL string `toml:"oauth_google_callback_url"`
	OAuthFinalRedirectURL string `toml:"oauth_final_redirect_url"`
	StorageEndpointURL string `toml:"storage_endpoint_url"`
	StorageRegion string `toml:"storage_region"`
	StorageKeyID string `toml:"storage_key_id"`
	StorageSecretKey string `toml:"storage_secret_key"`
	UsercontentBucket string `toml:"usercontent_bucket"`
}
