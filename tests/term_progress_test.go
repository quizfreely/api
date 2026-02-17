package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTermProgressLifecycle(t *testing.T) {
	// 1. Setup: user1 creates a public studyset with a term
	createSSBody := map[string]interface{}{
		"query": `mutation {
			createStudyset(studyset: {title: "Public Set for Progress", private: false}) { id }
		}`,
	}
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createSSBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ := http.DefaultClient.Do(req)
	var createSSResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createSSResult)
	studysetID := getNested(createSSResult, "data", "createStudyset", "id").(string)

	createTermsBody := map[string]interface{}{
		"query": `mutation CreateTerms($sid: ID!, $terms: [NewTermInput!]!) {
			createTerms(studysetId: $sid, terms: $terms) { id }
		}`,
		"variables": map[string]interface{}{
			"sid": studysetID,
			"terms": []map[string]interface{}{
				{"term": "T1", "def": "D1", "sortOrder": 0},
			},
		},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createTermsBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ = http.DefaultClient.Do(req)
	var createTermsResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createTermsResult)
	createTerms := getNested(createTermsResult, "data", "createTerms").([]interface{})
	termID := createTerms[0].(map[string]interface{})["id"].(string)

	// 2. user2 updates progress for user1's term
	updateProgressBody := map[string]interface{}{
		"query": `mutation UpdateProgress($input: [TermProgressInput!]!) {
			updateTermProgress(termProgress: $input) {
				termLeitnerSystemBox
				termCorrectCount
			}
		}`,
		"variables": map[string]interface{}{
			"input": []map[string]interface{}{
				{
					"termId":               termID,
					"termCorrectIncrease":  1,
					"termLeitnerSystemBox": 1,
				},
			},
		},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(updateProgressBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, _ = http.DefaultClient.Do(req)
	var updateResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&updateResult)
	if updateResult["errors"] != nil {
		t.Fatalf("updateTermProgress failed: %v", updateResult["errors"])
	}
	progress := getNested(updateResult, "data", "updateTermProgress").([]interface{})
	require.Equal(t, float64(1), progress[0].(map[string]interface{})["termCorrectCount"])

	// 3. User Isolation: Check that user1's progress for the same term is still empty/initial
	queryProgressBody := map[string]interface{}{
		"query": `query GetTerm($id: ID!) {
			term(id: $id) {
				progress {
					termCorrectCount
				}
			}
		}`,
		"variables": map[string]interface{}{
			"id": termID,
		},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(queryProgressBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ = http.DefaultClient.Do(req)
	var queryResult1 map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&queryResult1)
	require.Nil(t, getNested(queryResult1, "data", "term", "progress"), "user1 should have no progress for the term yet")

	// 4. Verification: user2 should see their own progress
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(queryProgressBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, _ = http.DefaultClient.Do(req)
	var queryResult2 map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&queryResult2)
	require.Equal(t, float64(1), getNested(queryResult2, "data", "term", "progress", "termCorrectCount"))

	// 5. Private Set Security: user1 marks set as private; user2 tries to update progress
	updateSSBody := map[string]interface{}{
		"query": `mutation UpdateSS($id: ID!) {
			updateStudyset(id: $id, studyset: {title: "Now Private", private: true}) { id }
		}`,
		"variables": map[string]interface{}{
			"id": studysetID,
		},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(updateSSBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	http.DefaultClient.Do(req)

	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(updateProgressBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, _ = http.DefaultClient.Do(req)
	var privateResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&privateResult)
	require.NotNil(t, privateResult["errors"], "user2 should NOT be able to update progress for a private term")
}
