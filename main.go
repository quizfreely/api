package main

import (
	"context"
	"net/http"
	"os"
	"quizfreely/api/server"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Port string `toml:"port"`
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

func main() {
	var config Config
	configBytes, err := os.ReadFile("config.toml")
	if err != nil {
		log.Fatal().Err(err).Msgf("Error reading config.toml")
	}
	err = toml.Unmarshal([]byte(configBytes), &config)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error parsing config.toml")
	}

	if config.PrettyLog {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{Out: os.Stderr},
		)
	} else {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}

	var dbPool *pgxpool.Pool
	dbPool, err = pgxpool.New(
		context.Background(),
		config.DBURL,
	)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error creating database pool")
	}
	defer dbPool.Close()

	router := server.NewRouter(dbPool, config)

	log.Info().Msg(
		"http://localhost:" + port + "/graphiql for GraphiQL",
	)
	log.Fatal().Err(
		http.ListenAndServe(":"+port, router),
	).Msgf("Error starting server")

	startSessionCleanupJob(dbPool)
}

func startSessionCleanupJob(dbPool *pgxpool.Pool) {
	ticker := time.NewTicker(24 * time.Hour) // Once per day
	go func() {
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			_, err := dbPool.Exec(ctx, "DELETE FROM auth.sessions WHERE expire_at < now()")
			if err != nil {
				log.Error().Err(err).Msg("Failed to clean up expired sessions")
			}
			cancel()
		}
	}()
}
