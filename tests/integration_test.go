package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"quizfreely/api/server"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startPostgres(t *testing.T) (tc.Container, string) {
	ctx := context.Background()

	req := tc.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	dbURL := "postgres://test:test@" + host + ":" + port.Port() + "/testdb?sslmode=disable"

	return container, dbURL
}

var testServer *httptest.Server
var container tc.Container
var dbURL string
var dbPool *pgxpool.Pool

var user1ID string
var user1Token string
var user2ID string
var user2Token string

func TestMain(m *testing.M) {
	ctx := context.Background()

	t := &testing.T{}
	container, dbURL = startPostgres(t)

	parsedDBURL, err := url.Parse(dbURL)
	if err != nil {
		panic(err)
	}

	dbPool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		panic(err)
	}

	_, err = dbPool.Exec(ctx, "CREATE ROLE quizfreely_api")
	if err != nil {
		panic(err)
	}

	dbMigration := dbmate.New(parsedDBURL)
	dbMigration.MigrationsDir = []string{"../db/migrations"}
	dbMigration.SchemaFile = "../db/schema.sql"
	dbMigration.AutoDumpSchema = false
	err = dbMigration.CreateAndMigrate()
	if err != nil {
		panic(err)
	}

	err = pgxscan.Get(
		ctx,
		dbPool,
		&user1ID,
		`INSERT INTO auth.users (username, encrypted_password, display_name, auth_type)
VALUES ($1, crypt($1, gen_salt('bf')), $1, 'USERNAME_PASSWORD')
RETURNING id`,
		"user1",
	)
	if err != nil {
		panic(err)
	}
	err = pgxscan.Get(
		ctx,
		dbPool,
		&user2ID,
		`INSERT INTO auth.users (username, encrypted_password, display_name, auth_type)
VALUES ($1, crypt($1, gen_salt('bf')), $1, 'USERNAME_PASSWORD')
RETURNING id`,
		"user2",
	)
	if err != nil {
		panic(err)
	}

	err = pgxscan.Get(
		ctx,
		dbPool,
		&user1Token,
		`INSERT INTO auth.sessions (user_id)
VALUES ($1) RETURNING token`,
		user1ID,
	)
	if err != nil {
		panic(err)
	}
	err = pgxscan.Get(
		ctx,
		dbPool,
		&user2Token,
		`INSERT INTO auth.sessions (user_id)
VALUES ($1) RETURNING token`,
		user2ID,
	)
	if err != nil {
		panic(err)
	}

	router := server.NewRouter(dbPool)
	testServer = httptest.NewServer(router)

	code := m.Run()

	testServer.Close()
	dbPool.Close()
	container.Terminate(ctx)

	os.Exit(code)
}

func marshal(v any) *bytes.Reader {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(b)
}

func getNested(m map[string]interface{}, keys ...string) interface{} {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			return current[key]
		}
		next, ok := current[key].(map[string]interface{})
		if !ok {
			return nil
		}
		current = next
	}
	return nil
}
