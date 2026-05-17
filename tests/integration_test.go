package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
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
var modUser1ID string
var modUser1Token string

func TestMain(m *testing.M) {
	code := func() int {
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
			&modUser1ID,
			`INSERT INTO auth.users (username, encrypted_password, display_name, auth_type, mod_perms)
VALUES ($1, crypt($1, gen_salt('bf')), $1, 'USERNAME_PASSWORD', true)
RETURNING id`,
			"modUser1",
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

		err = pgxscan.Get(
			ctx,
			dbPool,
			&modUser1Token,
			`INSERT INTO auth.sessions (user_id)
VALUES ($1) RETURNING token`,
			modUser1ID,
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

func TestStudysetSeoIndexingApproved(t *testing.T) {
	// 1. Create public studyset (by user1)
	createBody := map[string]interface{}{
		"query": `mutation CreateStudyset($input: StudysetInput!) {
			createStudyset(studyset: $input, draft: false) {
				id
				seoIndexingApproved
			}
		}`,
		"variables": map[string]interface{}{
			"input": map[string]interface{}{
				"title":   "Test SEO",
				"private": false,
			},
		},
	}

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var createResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResult)
	require.NoError(t, err)
	require.Nil(t, createResult["errors"])

	studysetID := getNested(createResult, "data", "createStudyset", "id").(string)
	seoIndexingApproved := getNested(createResult, "data", "createStudyset", "seoIndexingApproved").(bool)
	require.False(t, seoIndexingApproved, "Should be false by default")

	// 2. Random user (user2) attempts to set seo indexing
	setSeoBody := map[string]interface{}{
		"query": `mutation SetStudysetSeoIndexing($studysetId: ID!, $approved: Boolean!) {
			setStudysetSeoIndexing(studysetId: $studysetId, approved: $approved)
		}`,
		"variables": map[string]interface{}{
			"studysetId": studysetID,
			"approved":   true,
		},
	}

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(setSeoBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var errResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&errResult)
	require.NotNil(t, errResult["errors"], "user2 should be rejected")

	// 3. Owner (user1) attempts to set seo indexing
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(setSeoBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var errResult2 map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&errResult2)
	require.NotNil(t, errResult2["errors"], "owner should be rejected")

	// 4. Moderator (modUser1) sets seo indexing
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(setSeoBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+modUser1Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var successResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&successResult)
	require.Nil(t, successResult["errors"], "moderator should succeed")
	require.True(t, getNested(successResult, "data", "setStudysetSeoIndexing").(bool))

	// 5. Query to ensure it was updated
	queryBody := map[string]interface{}{
		"query": `query GetStudyset($id: ID!) {
			studyset(id: $id) {
				seoIndexingApproved
			}
		}`,
		"variables": map[string]interface{}{
			"id": studysetID,
		},
	}
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(queryBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var queryResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&queryResult)
	require.Nil(t, queryResult["errors"])
	require.True(t, getNested(queryResult, "data", "studyset", "seoIndexingApproved").(bool), "Should be true now")
}
