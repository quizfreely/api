package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPracticeTestLifecycle(t *testing.T) {
	// 1. Setup: user1 creates a public studyset
	createSSBody := map[string]interface{}{
		"query": `mutation {
			createStudyset(studyset: {title: "Public Set for PT", private: false}) { id }
		}`,
	}
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createSSBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ := http.DefaultClient.Do(req)
	var createSSResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createSSResult)
	studysetID := getNested(createSSResult, "data", "createStudyset", "id").(string)

	// 2. user2 records a practice test for user1's public studyset
	recordPTBody := map[string]interface{}{
		"query": `mutation RecordPT($input: PracticeTestInput!) {
			recordPracticeTest(input: $input) {
				id
				questionsCorrect
				questionsTotal
			}
		}`,
		"variables": map[string]interface{}{
			"input": map[string]interface{}{
				"studysetId":       studysetID,
				"questionsCorrect": 8,
				"questionsTotal":   10,
				"questions":        []interface{}{}, // simplified for test
			},
		},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(recordPTBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, _ = http.DefaultClient.Do(req)
	var recordResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&recordResult)
	require.Nil(t, recordResult["errors"], "user2 should be able to record PT for public set")
	ptID := getNested(recordResult, "data", "recordPracticeTest", "id").(string)

	// 3. user2 updates their own practice test
	updatePTBody := map[string]interface{}{
		"query": `mutation UpdatePT($input: PracticeTestInput!) {
			updatePracticeTest(input: $input) {
				id
				questionsCorrect
			}
		}`,
		"variables": map[string]interface{}{
			"input": map[string]interface{}{
				"id":               ptID,
				"questionsCorrect": 9,
				"questionsTotal":   10,
				"questions":        []interface{}{},
			},
		},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(updatePTBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, _ = http.DefaultClient.Do(req)
	var updateResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&updateResult)
	require.Nil(t, updateResult["errors"], "user2 should be able to update their own PT")
	require.Equal(t, float64(9), getNested(updateResult, "data", "updatePracticeTest", "questionsCorrect"))

	// 4. Invalid Authz: user1 tries to update user2's practice test
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(updatePTBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ = http.DefaultClient.Do(req)
	var unauthorizedResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&unauthorizedResult)
	require.NotNil(t, unauthorizedResult["errors"], "user1 should NOT be able to update user2's PT")

	// 5. Private Set Security: user1 creates a private studyset; user2 tries to record PT
	createPrivateSSBody := map[string]interface{}{
		"query": `mutation {
			createStudyset(studyset: {title: "Private Set", private: true}) { id }
		}`,
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createPrivateSSBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ = http.DefaultClient.Do(req)
	var createPrivateResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createPrivateResult)
	privateStudysetID := getNested(createPrivateResult, "data", "createStudyset", "id").(string)

	recordPTBody["variables"].(map[string]interface{})["input"].(map[string]interface{})["studysetId"] = privateStudysetID
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(recordPTBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, _ = http.DefaultClient.Do(req)
	var privateResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&privateResult)
	require.NotNil(t, privateResult["errors"], "user2 should NOT be able to record PT for user1's private set")
}
