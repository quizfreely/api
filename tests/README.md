# Integration Tests

Test cases covered by integration tests in `quizfreely/api/tests` are listed here.

## `integration_test.go`

`tests/integration_test.go` sets up the database and API for the tests. PostgreSQL runs in a container using Testcontainers for Go, and quizfreely-api "runs itself" using the same code from `server/server.go` (without `main.go` because env/DB related stuff in `main.go` is instead done by `integration_test.go` for tests).

This setup process in the `TestMain` function in `integration_test.go` is only run once for the whole package, so all the tests use the same database container and aren't isolated, which is fine for these kinds of integration tests and keeps everything efficient.

- **TestMain**:
    - Starts a PostgreSQL container.
    - Runs database migrations.
    - Creates test users (`user1`, `user2`) and sessions.
    - Starts the test HTTP server.
    - Runs all the tests (`m.Run()`).
    - Cleans up resources (server, database connection, container).

## `practice_test_test.go`

- **TestPracticeTestLifecycle**:
    1. **Setup**: `user1` creates a public studyset.
    2. **Record PT**: `user2` records a practice test for `user1`'s public studyset.
    3. **Update PT**: `user2` updates their own practice test.
    4. **Invalid Authz**: `user1` attempts to update `user2`'s practice test (should fail).
    5. **Private Set Security**: `user1` creates a private studyset; `user2` attempts to record a practice test for it (should fail).

## `progress_test.go`

- **TestTermProgressLifecycle**:
    1. **Setup**: `user1` creates a public studyset with a term.
    2. **Update Progress**: `user2` updates progress for `user1`'s term.
    3. **User Isolation**: Verifies that `user1`'s progress for the same term remains empty/initial (progress is per-user).
    4. **Verification**: Verifies that `user2` can see their own updated progress.
    5. **Private Set Security**: `user1` updates the studyset to be private; `user2` attempts to update progress for the term (should fail).

## `studyset_test.go`
Tests related to studyset CRUD operations.

- **TestStudysetLifecycle**:
    1. **Create Studyset**: `user1` creates a studyset (Valid Auth).
    2. **Update Studyset**: `user1` updates the studyset (Valid Auth).
    3. **Unauthorized Update**: `user2` attempts to update `user1`'s studyset (should fail).
    4. **Delete Studyset**: `user1` deletes the studyset (Valid Auth).
    5. **Unauthorized Delete**: `user2` attempts to delete `user1`'s studyset (should fail).

- **TestStudysetNoAuth**:
    1. **No Auth Creation**: anonymous user attempts to create a studyset (should fail).
    2. **No Auth Update**: anonymous user attempts to update a studyset (should fail).
    3. **No Auth Delete**: anonymous user attempts to delete a studyset (should fail).

## `term_test.go`
Tests related to term CRUD operations within a studyset.

- **TestTermLifecycle**:
    1. **Setup**: `user1` creates a studyset.
    2. **Create Terms**: `user1` adds terms to the studyset.
    3. **Update Terms**: `user1` updates terms in the studyset.
    4. **Unauthorized Term Modification**: `user2` attempts to add terms to `user1`'s studyset (should fail).
    5. **Delete Terms**: `user1` deletes terms from the studyset.
    6. **Unauthorized Edit**: `user2` attempts to edit terms in `user1`'s studyset (should fail).
    7. **Unauthorized Delete**: `user2` attempts to delete terms from `user1`'s studyset (should fail).

- **TestTermNoAuth**:
    1. **No Auth Creation**: anonymous user attempts to create terms (should fail).
    2. **No Auth Edit**: anonymous user attempts to edit terms (should fail).
    3. **No Auth Delete**: anonymous user attempts to delete terms (should fail).
