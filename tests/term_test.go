package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTermLifecycle(t *testing.T) {
	// 1. Setup: Create a studyset for user1
	createStudysetBody := map[string]interface{}{
		"query": `mutation {
			createStudyset(studyset: {title: "Term Test Set", private: true}) { id }
		}`,
	}
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createStudysetBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ := http.DefaultClient.Do(req)
	var createSSResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createSSResult)
	studysetID := getNested(createSSResult, "data", "createStudyset", "id").(string)

	// 2. Create Terms (Valid Auth - user1)
	createTermsBody := map[string]interface{}{
		"query": `mutation CreateTerms($studysetId: ID!, $terms: [NewTermInput!]!) {
			createTerms(studysetId: $studysetId, terms: $terms) {
				id
				term
				def
			}
		}`,
		"variables": map[string]interface{}{
			"studysetId": studysetID,
			"terms": []map[string]interface{}{
				{"term": "T1", "def": "D1", "sortOrder": 0},
				{"term": "T2", "def": "D2", "sortOrder": 1},
			},
		},
	}

	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createTermsBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)

	resp, _ = http.DefaultClient.Do(req)
	var createTermsResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createTermsResult)
	require.Nil(t, createTermsResult["errors"], "should have no errors on term creation")

	terms := getNested(createTermsResult, "data", "createTerms").([]interface{})
	require.Len(t, terms, 2)
	term1ID := terms[0].(map[string]interface{})["id"].(string)

	// 3. Update Terms (Valid Auth - user1)
	updateTermsBody := map[string]interface{}{
		"query": `mutation UpdateTerms($studysetId: ID!, $terms: [TermInput!]!) {
			updateTerms(studysetId: $studysetId, terms: $terms) {
				id
				term
			}
		}`,
		"variables": map[string]interface{}{
			"studysetId": studysetID,
			"terms": []map[string]interface{}{
				{"id": term1ID, "term": "T1 Updated", "def": "D1 Updated", "sortOrder": 0},
			},
		},
	}

	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(updateTermsBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)

	resp, _ = http.DefaultClient.Do(req)
	var updateTermsResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&updateTermsResult)
	require.Nil(t, updateTermsResult["errors"], "should have no errors on term update")

	updatedTerms := getNested(updateTermsResult, "data", "updateTerms").([]interface{})
	require.Equal(t, "T1 Updated", updatedTerms[0].(map[string]interface{})["term"])

	// 4. Unauthorized Term Modification (user2 trying to add terms to user1's studyset)
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createTermsBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)

	resp, _ = http.DefaultClient.Do(req)
	var unauthorizedResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&unauthorizedResult)
	require.NotNil(t, unauthorizedResult["errors"], "user2 should not be able to add terms to user1's studyset")

	// 5. Delete Terms (Valid Auth - user1)
	deleteTermsBody := map[string]interface{}{
		"query": `mutation DeleteTerms($studysetId: ID!, $ids: [ID!]!) {
			deleteTerms(studysetId: $studysetId, ids: $ids)
		}`,
		"variables": map[string]interface{}{
			"studysetId": studysetID,
			"ids":        []string{term1ID},
		},
	}

	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(deleteTermsBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)

	resp, _ = http.DefaultClient.Do(req)
	var deleteTermsResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&deleteTermsResult)
	require.Nil(t, deleteTermsResult["errors"], "should have no errors on term deletion")
	deletedIDs := getNested(deleteTermsResult, "data", "deleteTerms").([]interface{})
	require.Contains(t, deletedIDs, term1ID)
}

func TestTermNoAuth(t *testing.T) {
	body := map[string]interface{}{
		"query": `mutation {
			createTerms(studysetId: "123", terms: [{term: "X", def: "Y", sortOrder: 0}]) { id }
		}`,
	}

	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := http.DefaultClient.Do(req)
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	require.NotNil(t, result["errors"], "should fail without auth")
}
