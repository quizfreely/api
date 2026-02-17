package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveStudysetLifecycle(t *testing.T) {
	// 1. Setup: Create Studyset (user1)
	createBody := map[string]interface{}{
		"query": `mutation CreateStudyset($input: StudysetInput!) {
			createStudyset(studyset: $input) { id }
		}`,
		"variables": map[string]interface{}{
			"input": map[string]interface{}{"title": "Save Test Studyset", "private": false},
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
	studysetID := getNested(createResult, "data", "createStudyset", "id").(string)

	// 2. Save Studyset (user2)
	saveBody := map[string]interface{}{
		"query": `mutation SaveStudyset($studysetId: ID!) {
			saveStudyset(studysetId: $studysetId)
		}`,
		"variables": map[string]interface{}{
			"studysetId": studysetID,
		},
	}
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(saveBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var saveResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&saveResult)
	require.NoError(t, err)
	require.Nil(t, saveResult["errors"], "should save successfully")
	require.True(t, getNested(saveResult, "data", "saveStudyset").(bool))

	// 3. Verify Saved Status (user2)
	queryBody := map[string]interface{}{
		"query": `query GetStudyset($id: ID!) {
			studyset(id: $id) {
				id
				saved
			}
		}`,
		"variables": map[string]interface{}{"id": studysetID},
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
	require.Equal(t, true, getNested(queryResult, "data", "studyset", "saved"))

	// 4. Verify MySavedStudysets (user2)
	mySavedBody := map[string]interface{}{
		"query": `query MySavedStudysets {
			mySavedStudysets {
				edges {
					node {
						id
						title
					}
				}
			}
		}`,
	}
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(mySavedBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var mySavedResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&mySavedResult)
	require.NoError(t, err)
	edges := getNested(mySavedResult, "data", "mySavedStudysets", "edges").([]interface{})
	found := false
	for _, edge := range edges {
		node := edge.(map[string]interface{})["node"].(map[string]interface{})
		if node["id"] == studysetID {
			found = true
			break
		}
	}
	require.True(t, found, "saved studyset should be in mySavedStudysets")

	// 5. Unsave Studyset (user2)
	unsaveBody := map[string]interface{}{
		"query": `mutation UnsaveStudyset($studysetId: ID!) {
			unsaveStudyset(studysetId: $studysetId)
		}`,
		"variables": map[string]interface{}{
			"studysetId": studysetID,
		},
	}
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(unsaveBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var unsaveResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&unsaveResult)
	require.NoError(t, err)
	require.Nil(t, unsaveResult["errors"], "should unsave successfully")
	require.True(t, getNested(unsaveResult, "data", "unsaveStudyset").(bool))

	// 6. Verify Unsaved Status (user2)
	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(queryBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&queryResult)
	require.Equal(t, false, getNested(queryResult, "data", "studyset", "saved"))
}

func TestSaveNoAuth(t *testing.T) {
	saveBody := map[string]interface{}{
		"query": `mutation {
			saveStudyset(studysetId: "123")
		}`,
	}
	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(saveBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	require.NotNil(t, result["errors"], "should fail without auth")
}
