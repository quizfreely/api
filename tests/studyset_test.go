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
			createStudyset(studyset: $input) {
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
			updateStudyset(id: $id, studyset: $input) {
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
}

func TestStudysetNoAuth(t *testing.T) {
	createBody := map[string]interface{}{
		"query": `mutation {
			createStudyset(studyset: {title: "No Auth", private: false}) { id }
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
}
