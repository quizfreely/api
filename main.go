package main

import (
	"context"
	"net/http"
	"os"
	"quizfreely/api/server"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const defaultPort = "8008"

func main() {
	_ = godotenv.Load()
	/* godotenv means go dotenv, not godot env*/

	if os.Getenv("PRETTY_LOG") == "true" {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{Out: os.Stderr},
		)
	} else {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal().Msgf(
			`DB_URL is not set
Copy .env.example to .env and/or
check your environment variables`,
		)
	}

	var err error
	var dbPool *pgxpool.Pool
	dbPool, err = pgxpool.New(
		context.Background(),
		dbUrl,
	)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error creating database pool")
	}
	defer dbPool.Close()

	router := server.NewRouter(dbPool)

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
