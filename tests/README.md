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

## `term_progress_test.go`

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

## `folder_test.go`
Tests related to folder CRUD operations and organization.

- **TestFolderLifecycle**:
    1. **Create Folder**: `user1` creates a folder (Valid Auth).
    2. **Rename Folder**: `user1` renames the folder (Valid Auth).
    3. **Unauthorized Rename**: `user2` attempts to rename `user1`'s folder (should fail).
    4. **Query Folder**: `user1` queries the folder details.
    5. **Delete Folder**: `user1` deletes the folder (Valid Auth).
    6. **Unauthorized Delete**: `user2` attempts to delete `user1`'s folder (should fail). (Note: Verifies that an attempt to delete an unauthorized folder returns an error).

- **TestStudysetFolderOperations**:
    1. **Setup**: `user1` creates a studyset and a folder.
    2. **Set Folder**: `user1` assigns the studyset to the folder.
    3. **Verify Association**: Verifies the studyset links to the folder in queries.
    4. **Remove from Folder**: `user1` removes the studyset from the folder.
    5. **Verify Removal**: Verifies the studyset no longer links to the folder.

- **TestFolderNoAuth**:
    1. **No Auth Creation**: anonymous user attempts to create a folder (should fail).

- **TestSharedStudysetFolderOperations**:
    1. **Setup**: `user1` creates a public studyset.
    2. **Setup**: `user2` creates a folder.
    3. **Add Shared**: `user2` adds `user1`'s public studyset to `user2`'s folder.
    4. **Verify**: Verifies the studyset appears in `user2`'s folder.
    5. **Remove Shared**: `user2` removes the studyset from the folder.
    6. **Verify Removal**: Verifies the studyset is removed.

## `saved_studyset_test.go`
Tests related to saving studysets (bookmarks).

- **TestSaveStudysetLifecycle**:
    1. **Setup**: `user1` creates a studyset.
    2. **Save Studyset**: `user2` saves `user1`'s studyset.
    3. **Verify Saved**: Verifies the studyset is marked as saved for `user2`.
    4. **Verify MySavedStudysets**: Verifies the studyset appears in `user2`'s saved list.
    5. **Unsave Studyset**: `user2` removes the studyset from saved.
    6. **Verify Unsaved**: Verifies the studyset is no longer marked as saved.

- **TestSaveNoAuth**:
    1. **No Auth Save**: anonymous user attempts to save a studyset (should fail).
