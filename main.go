package main

import (
	"context"
	"net/http"
	"os"
	"time"
	"strconv"

	"quizfreely/api/server"
	qzfrAPIConfig "quizfreely/api/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	var config qzfrAPIConfig.Config
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
