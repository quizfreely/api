package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"quizfreely/api/config"
	"quizfreely/api/server"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startPostgres(t *testing.T) (tc.Container, string, string) {
	ctx := context.Background()

	req := tc.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "quizfreely_db_admin",
			"POSTGRES_PASSWORD": "testsAdminPassword",
			"POSTGRES_DB":       "quizfreely_test_db",
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

	adminDBURL := "postgres://quizfreely_db_admin:testsAdminPassword@" + host + ":" + port.Port() + "/quizfreely_test_db?sslmode=disable"
	apiDBURL := "postgres://quizfreely_api:testsAPIPassword@" + host + ":" + port.Port() + "/quizfreely_test_db?sslmode=disable"

	return container, adminDBURL, apiDBURL
}

var testServer *httptest.Server
var container tc.Container
var dbPool *pgxpool.Pool

var user1ID string
var user1Token string
var user2ID string
var user2Token string

func TestMain(m *testing.M) {
	code := func () int {
		// NOTE: this immediately invoked func/IIFE is used because `defer` needs the function to return BEFORE os.Exit is called
		// os.Exit will skip stuff with `defer`, but using `defer` is much better than duplicaing `Close`/`Terminate` calls before every `panic`

		ctx := context.Background()

		t := &testing.T{}
		container, adminDBURL, apiDBURL := startPostgres(t)
		defer container.Terminate(ctx)

		parsedAdminDBURL, err := url.Parse(adminDBURL)
		if err != nil {
			panic(err)
		}

		adminConn, err := pgx.Connect(ctx, adminDBURL)
		if err != nil {
			panic(err)
		}
		_, err = adminConn.Exec(ctx, "CREATE ROLE quizfreely_api LOGIN PASSWORD 'testsAPIPassword'")
		if err != nil {
			adminConn.Close(ctx)
			panic(err)
		}
		adminConn.Close(ctx)

		dbMigration := dbmate.New(parsedAdminDBURL)
		dbMigration.MigrationsDir = []string{"../db/migrations"}
		dbMigration.SchemaFile = "../db/schema.sql"
		dbMigration.AutoDumpSchema = false
		err = dbMigration.CreateAndMigrate()
		if err != nil {
			panic(err)
		}

		dbPool, err = pgxpool.New(ctx, apiDBURL)
		if err != nil {
			panic(err)
		}
		defer dbPool.Close()

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

		router := server.NewRouter(
			config.Config{
				BasePath:          "/",
				EnableOAuthGoogle: false,
			},
			dbPool,
			nil,
		)
		testServer = httptest.NewServer(router)
		defer testServer.Close()

		return m.Run() /* return exit code */
	}()
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
