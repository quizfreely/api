package config

type Config struct {
	Port                     int    `toml:"port"`
	DBURL                    string `toml:"db_url"`
	PrettyLog                bool   `toml:"pretty_log"`
	BasePath                 string `toml:"base_path"`
	EnableOAuthGoogle        bool   `toml:"enable_oauth_google"`
	OAuthGoogleClientID      string `toml:"oauth_google_client_id"`
	OAuthGoogleClientSecret  string `toml:"oauth_google_client_secret"`
	OAuthGoogleCallbackURL   string `toml:"oauth_google_callback_url"`
	OAuthFinalRedirectURL    string `toml:"oauth_final_redirect_url"`
	StorageEndpointURL       string `toml:"storage_endpoint_url"`
	StorageRegion            string `toml:"storage_region"`
	StorageKeyID             string `toml:"storage_key_id"`
	StorageSecretKey         string `toml:"storage_secret_key"`
	UsercontentBucket        string `toml:"usercontent_bucket"`
	UsercontentBaseURL       string `toml:"usercontent_base_url"`
	SessionCleanupCronSpec   string `toml:"session_cleanup_cron_spec"`
	TermImageCleanupCronSpec string `toml:"term_image_cleanup_cron_spec"`
	EnableWebImport          bool   `toml:"enable_web_import"`
	UseCrawlbase             bool   `toml:"use_crawlbase"`
	CrawlbaseAPIKey               string `toml:"crawlbase_api_key"`
	UseZyte                  bool   `toml:"use_zyte"`
	ZyteAPIKey               string `toml:"zyte_api_key"`
	TryZyteBeforeCrawlbase                  bool   `toml:"try_zyte_before_crawlbase"`
}
