package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	qzfrAPIConfig "quizfreely/api/config"
	"quizfreely/api/server"
	"quizfreely/api/storage"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	var config qzfrAPIConfig.Config
	configFile, err := os.Open("config.toml")
	if err != nil {
		fmt.Printf("❌ Error opening/reading config.toml: %v\n", err)
		os.Exit(1)
	}
	err = toml.NewDecoder(configFile).DisallowUnknownFields().Decode(&config)
	if err != nil {
		var decodeErr *toml.DecodeError
		var strictMissingErr *toml.StrictMissingError
		if errors.As(err, &decodeErr) {
			fmt.Printf("❌ Error in config.toml\n%s\n", decodeErr.String())
		} else if errors.As(err, &strictMissingErr) {
			fmt.Printf("❌ Error in config.toml: Unknown field/setting that doesn't exist\n%s\n", strictMissingErr.String())
		} else {
			fmt.Printf("❌ Error decoding/parsing config.toml: %v\n", err)
		}
		os.Exit(1)
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

	var s3Client *s3.Client
	if config.StorageEndpointURL != "" {
		s3Client = storage.InitS3Client(
			config.StorageEndpointURL,
			config.StorageRegion,
			config.StorageKeyID,
			config.StorageSecretKey,
		)
	}

	router := server.NewRouter(config, dbPool, s3Client)

	port := strconv.Itoa(config.Port)
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
