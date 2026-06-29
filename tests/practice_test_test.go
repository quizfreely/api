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
			createStudyset(studyset: {title: "Public Set for PT", private: false}, draft: false) { id }
		}`,
	}
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createSSBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ := http.DefaultClient.Do(req)
	var createSSResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createSSResult)
	studysetID := getNested(createSSResult, "data", "createStudyset", "id").(string)

	// 2. Setup: Add terms to user1's studyset
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
	terms := getNested(createTermsResult, "data", "createTerms").([]interface{})
	term1ID := terms[0].(map[string]interface{})["id"].(string)
	term1Text := terms[0].(map[string]interface{})["term"].(string)
	term1Def := terms[0].(map[string]interface{})["def"].(string)

	// 3. user2 records a practice test for user1's public studyset
	recordPTBody := map[string]interface{}{
		"query": `mutation RecordPT($input: PracticeTestInput!) {
			recordPracticeTest(input: $input) {
				id
				questionsCorrect
				questionsTotal
				questions {
					id
				}
			}
		}`,
		"variables": map[string]interface{}{
			"input": map[string]interface{}{
				"questions": []interface{}{
					map[string]interface{}{
						"frq": map[string]interface{}{
							"term": map[string]interface{}{
								"id":  term1ID,
								"term": term1Text,
								"def":  term1Def,
							},
							"answerWith":     "DEF",
							"correct":        true,
							"answeredString": term1Def,
						},
					},
				},
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
	q1ID := getNested(recordResult, "data", "recordPracticeTest", "questions", 0, "id").(string)

	// 4. user2 updates their own practice test question
	updatePTQBody := map[string]interface{}{
		"query": `mutation UpdatePTQ($id: ID!, $correct: Boolean!) {
			updatePracticeTestQuestion(id: $id, correct: $correct) {
				frq {
					correct
				}
			}
		}`,
		"variables": map[string]interface{}{
			"id":      q1ID,
			"correct": false,
		},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(updatePTQBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, _ = http.DefaultClient.Do(req)
	var updateResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&updateResult)
	require.Nil(t, updateResult["errors"], "user2 should be able to update their own PTQ")
	require.False(t, getNested(updateResult, "data", "updatePracticeTestQuestion", "frq", "correct").(bool))

	// 5. Invalid Authz: user1 tries to update user2's practice test question
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(updatePTQBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ = http.DefaultClient.Do(req)
	var unauthorizedResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&unauthorizedResult)
	require.NotNil(t, unauthorizedResult["errors"], "user1 should NOT be able to update user2's PTQ")

	// 6. Private Set Security (Implicit): user2 tries to record PT for a term in a private studyset

	// Create private set and term
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(map[string]interface{}{
		"query": `mutation {
			createStudyset(studyset: {title: "Private Set", private: true}, draft: false) { id }
		}`,
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ = http.DefaultClient.Do(req)
	var createSSResult2 map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createSSResult2)
	privateStudysetID := getNested(createSSResult2, "data", "createStudyset", "id").(string)

	createTermsBody["variables"].(map[string]interface{})["studysetId"] = privateStudysetID
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(createTermsBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user1Token)
	resp, _ = http.DefaultClient.Do(req)
	var createTermsResult2 map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createTermsResult2)
	privateTerms := getNested(createTermsResult2, "data", "createTerms").([]interface{})
	privateTermID := privateTerms[0].(map[string]interface{})["id"].(string)

	recordPTBody["variables"].(map[string]interface{})["input"].(map[string]interface{})["questions"] = []interface{}{
		map[string]interface{}{
			"mcq": map[string]interface{}{
				"term": map[string]interface{}{
					"id":    privateTermID,
					"term":  "X",
					"def":   "Y",
				},
				"answerWith":         "DEF",
				"correct":            true,
				"correctChoiceIndex": 0,
				"answeredIndex":      0,
				"distractors":        []interface{}{},
			},
		},
	}
	req, _ = http.NewRequest(http.MethodPost, testServer.URL+"/graphql", marshal(recordPTBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+user2Token)
	resp, _ = http.DefaultClient.Do(req)
	var privateResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&privateResult)
	require.NotNil(t, privateResult["errors"], "user2 should NOT be able to record PT for user1's private set")
}
