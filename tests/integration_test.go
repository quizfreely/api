package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"quizfreely/api/server"

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

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, dbURL := startPostgres(&testing.T{})

	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		panic(err)
	}

	// TODO: run migrations here

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

func TestUnauthorizedGraphQL(t *testing.T) {
	query := `{
		query {
			authed
		}
	}`

	body := map[string]string{
		"query": query,
	}

	resp, err := http.Post(
		testServer.URL+"/graphql",
		"application/json",
		marshal(body),
	)
	require.NoError(t, err)

	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	require.NotNil(t, result["errors"])
}
