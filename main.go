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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pelletier/go-toml/v2"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
			zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04"},
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
		log.Fatal().Err(err).Msg("Error creating database pool")
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

	c := cron.New()
	c.AddFunc(config.SessionCleanupCronSpec, func() {
		sessionCleanupJob(dbPool)
	})
	c.AddFunc(config.TermImageCleanupCronSpec, func() {
		termImageCleanupJob(dbPool, s3Client, config.UsercontentBucket)
	})
	c.Start()
	/* start cron jobs BEFORE starting server because http.ListenAndServe (below) is blocking */

	port := strconv.Itoa(config.Port)
	log.Info().Msg(
		"http://localhost:" + port + "/graphiql for GraphiQL",
	)
	log.Fatal().Err(
		http.ListenAndServe(":"+port, router),
	).Msg("Error starting server")
}

func sessionCleanupJob(dbPool *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	log.Info().Msg("Running sessionCleanupJob")
	_, err := dbPool.Exec(ctx, "DELETE FROM auth.sessions WHERE expire_at < now()")
	if err != nil {
		log.Error().Err(err).Msg("Failed to clean up expired sessions")
	}
}

func termImageCleanupJob(dbPool *pgxpool.Pool, storage *s3.Client, usercontentBucket string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	log.Info().Msg("Running termImageCleanupJob")
	var keys []string
	err := pgxscan.Select(
		ctx,
		dbPool,
		&keys,
		`SELECT object_key
		FROM images i
		WHERE NOT EXISTS(
			SELECT 1 FROM terms t
			WHERE t.term_image_key = i.object_key OR t.def_image_key = i.object_key
		)`,
	)
	if err != nil {
		log.Error().Err(err).Msg("DB err while getting keys to clean up term images")
	}

	objects := make([]types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = types.ObjectIdentifier{
			Key: aws.String(key),
		}
	}

	for i := 0; i < len(keys); i += 1000 {
		end := i + 1000
		if end > len(keys) {
			end = len(keys)
		}

		output, err := storage.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: &usercontentBucket,
			Delete: &types.Delete{
				Objects: objects[i:end],
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			log.Error().Err(err).Msg("error deleting term images from S3 in term image cleanup job")
			continue
		}

		deletedKeys := make([]string, 0, len(output.Deleted))
		for _, deletedObject := range output.Deleted {
			if deletedObject.Key == nil {
				continue
			}

			deletedKeys = append(deletedKeys, *deletedObject.Key)
		}

		if len(deletedKeys) > 0 {
			_, err = dbPool.Exec(
				ctx,
				"DELETE FROM images WHERE object_key = ANY($1)",
				deletedKeys,
			)
			if err != nil {
				log.Error().Err(err).Msg("DB err deleting rows from term_images after successful delete from S3")
			}
		}
	}
}
