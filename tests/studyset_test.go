package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStudysetLifecycle(t *testing.T) {
	// 1. Create Studyset (Valid Auth - user1)
	createBody := map[string]interface{}{
		"query": `mutation CreateStudyset($input: StudysetInput!) {
			createStudyset(studyset: $input, draft: false) {
				id
				title
				private
			}
		}`,
		"variables": map[string]interface{}{
			"input": map[string]interface{}{
				"title":   "Test Studyset",
				"private": true,
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
	require.Nil(t, createResult["errors"], "should have no errors on creation: %v", createResult["errors"])

	studysetID := getNested(createResult, "data", "createStudyset", "id").(string)
	require.NotEmpty(t, studysetID)
	require.Equal(t, "Test Studyset", getNested(createResult, "data", "createStudyset", "title"))

	// 2. Update Studyset (Valid Auth - user1)
	updateBody := map[string]interface{}{
		"query": `mutation UpdateStudyset($id: ID!, $input: StudysetInput!) {
			updateStudyset(id: $id, studyset: $input, draft: false) {
				id
				title
			}
		}`,
		"variables": map[string]interface{}{
			"id": studysetID,
			"input": map[string]interface{}{
				"title":   "Updated Title",
				"private": false,
			},
		},
	}

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(updateBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var updateResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&updateResult)
	require.NoError(t, err)
	require.Nil(t, updateResult["errors"], "should have no errors on update")
	require.Equal(t, "Updated Title", getNested(updateResult, "data", "updateStudyset", "title"))

	// 3. Unauthorized Update (user2 trying to update user1's studyset)
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(updateBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var unauthorizedResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&unauthorizedResult)
	require.NoError(t, err)
	require.NotNil(t, unauthorizedResult["errors"], "should return authorization error")

	// 4. Delete Studyset (Valid Auth - user1)
	deleteBody := map[string]interface{}{
		"query": `mutation DeleteStudyset($id: ID!) {
			deleteStudyset(id: $id)
		}`,
		"variables": map[string]interface{}{
			"id": studysetID,
		},
	}

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(deleteBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var deleteResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&deleteResult)
	require.NoError(t, err)
	require.Nil(t, deleteResult["errors"], "should have no errors on deletion")
	require.Equal(t, studysetID, getNested(deleteResult, "data", "deleteStudyset").(string))
	// 5. Unauthorized Delete (user2 trying to delete user1's studyset)
	// First, recreate a studyset since we deleted one
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	var recreateResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&recreateResult)
	newStudysetID := getNested(recreateResult, "data", "createStudyset", "id").(string)

	deleteBody["variables"].(map[string]interface{})["id"] = newStudysetID
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(deleteBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	var unauthorizedDeleteResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&unauthorizedDeleteResult)
	require.NotNil(t, unauthorizedDeleteResult["errors"], "should return authorization error on delete")
}

func TestStudysetNoAuth(t *testing.T) {
	createBody := map[string]interface{}{
		"query": `mutation {
			createStudyset(studyset: {title: "No Auth", private: false}, draft: false) { id }
		}`,
	}

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.NotNil(t, result["errors"], "should fail without auth")

	// Unauthenticated Edit
	updateBody := map[string]interface{}{
		"query": `mutation {
			updateStudyset(id: "123", studyset: {title: "No Auth", private: false}, draft: false) { id }
		}`,
	}
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(updateBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	var updateResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&updateResult)
	require.NotNil(t, updateResult["errors"], "should fail update without auth")

	// Unauthenticated Delete
	deleteBody := map[string]interface{}{
		"query": `mutation {
			deleteStudyset(id: "123")
		}`,
	}
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(deleteBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	var deleteResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&deleteResult)
	require.NotNil(t, deleteResult["errors"], "should fail delete without auth")
}

func TestDraftStudysetLifecycle(t *testing.T) {
	// 1. user1 creates a draft studyset with empty title
	createBody := map[string]interface{}{
		"query": `mutation CreateStudyset($input: StudysetInput!) {
			createStudyset(studyset: $input, draft: true) {
				id
				title
				draft
			}
		}`,
		"variables": map[string]interface{}{
			"input": map[string]interface{}{
				"title":   "",
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
	require.Nil(t, createResult["errors"], "should have no errors on draft creation")

	studysetID := getNested(createResult, "data", "createStudyset", "id").(string)
	require.NotEmpty(t, studysetID)
	require.Equal(t, "", getNested(createResult, "data", "createStudyset", "title"))
	require.Equal(t, true, getNested(createResult, "data", "createStudyset", "draft"))

	// 2. user2 attempts to view the draft studyset (should fail / not found)
	queryBody := map[string]interface{}{
		"query": `query GetStudyset($id: ID!) {
			studyset(id: $id) {
				id
				title
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
	err = json.NewDecoder(resp.Body).Decode(&queryResult)
	require.NoError(t, err)
	require.NotNil(t, queryResult["errors"], "should return error when attempting to view someone else's draft")

	// 3. user1 updates the studyset to making it no longer a draft and sets a title
	updateBody := map[string]interface{}{
		"query": `mutation UpdateStudyset($id: ID!, $input: StudysetInput!) {
			updateStudyset(id: $id, studyset: $input, draft: false) {
				id
				title
				draft
			}
		}`,
		"variables": map[string]interface{}{
			"id": studysetID,
			"input": map[string]interface{}{
				"title":   "Published Set",
				"private": false,
			},
		},
	}
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(updateBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var updateResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&updateResult)
	require.NoError(t, err)
	require.Nil(t, updateResult["errors"], "should have no errors on update")
	require.Equal(t, "Published Set", getNested(updateResult, "data", "updateStudyset", "title"))
	require.Equal(t, false, getNested(updateResult, "data", "updateStudyset", "draft"))

	// 4. user2 attempts to view the published studyset (should succeed)
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(queryBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var queryResult2 map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&queryResult2)
	require.NoError(t, err)
	require.Nil(t, queryResult2["errors"], "should have no errors viewing published studyset")
	require.Equal(t, "Published Set", getNested(queryResult2, "data", "studyset", "title"))
}
